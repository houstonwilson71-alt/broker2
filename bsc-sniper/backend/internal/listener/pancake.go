package listener

import (
	"context"
	"encoding/json"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	redisclient "github.com/bsc-sniper/backend/internal/redis"
	"go.uber.org/zap"
)

// ─── Factory addresses ────────────────────────────────────────────────────────

const (
	FactoryV2     = "0xcA143Ce32Fe78f1f7019d7d551a6402fC5350c73"
	FactoryV3     = "0x0BFbCF9fa4f9C56B0F40a671Ad40E0805A091865"
	FactoryStable = "0x25a55f9f2279A54951133D503490342b50E5cd15"
)

// quoteTokens maps lowercase BSC addresses to their symbol.
// A pair is accepted if exactly one token is in this whitelist;
// that token is the quote and the other token is the meme coin.
var quoteTokens = map[string]string{
	"0xbb4cdb9cbd36b01bd1cbaebf2de08d9173bc095c": "WBNB",
	"0x55d398326f99059ff775485246999027b3197955": "USDT",
	"0xe9e7cea3dedca5984780bafc599bd69add087d56": "BUSD",
	"0x8ac76a51cc950d9822d68b83fe1ad97b32cd580d": "USDC",
	"0x2170ed0880ac9a755fd29b2688956bd959f933f8": "ETH",
	"0x0e09fabb73bd3ade0a17ecc321fd13a19e81ce82": "CAKE",
}

// ─── Event ABIs ───────────────────────────────────────────────────────────────

const pairCreatedABIJSON = `[{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"token0","type":"address"},{"indexed":true,"internalType":"address","name":"token1","type":"address"},{"indexed":false,"internalType":"address","name":"pair","type":"address"},{"indexed":false,"internalType":"uint256","name":"allPairsLength","type":"uint256"}],"name":"PairCreated","type":"event"}]`

const poolCreatedABIJSON = `[{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"token0","type":"address"},{"indexed":true,"internalType":"address","name":"token1","type":"address"},{"indexed":true,"internalType":"uint24","name":"fee","type":"uint24"},{"indexed":false,"internalType":"int24","name":"tickSpacing","type":"int24"},{"indexed":false,"internalType":"address","name":"pool","type":"address"}],"name":"PoolCreated","type":"event"}]`

const newStablePairABIJSON = `[{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"token0","type":"address"},{"indexed":true,"internalType":"address","name":"token1","type":"address"},{"indexed":false,"internalType":"address","name":"pairContract","type":"address"},{"indexed":false,"internalType":"uint256","name":"A","type":"uint256"}],"name":"NewStableSwapPair","type":"event"}]`

// ─── Domain types ─────────────────────────────────────────────────────────────

type NewPairEvent struct {
	Token0      string `json:"token0"`
	Token1      string `json:"token1"`
	PairAddress string `json:"pair_address"`
	MemeToken   string `json:"meme_token"`
	QuoteToken  string `json:"quote_token"` // canonical-case address of the quote token
	QuoteSymbol string `json:"quote_symbol"` // e.g. "WBNB", "USDT"
	Source      string `json:"source"`      // "v2" | "v3" | "stable"
	FeeTier     uint32 `json:"fee_tier"`    // V3 fee tier in hundredths of a bip (500, 2500, 10000); 0 for V2/Stable
	BlockNumber uint64 `json:"block_number"`
	Timestamp   int64  `json:"timestamp"`
}

// ─── Ring-buffer deduplicator ─────────────────────────────────────────────────

const ringSize = 10_000

type ringBuffer struct {
	mu   sync.Mutex
	seen map[string]struct{}
	ring [ringSize]string
	pos  int
}

func (rb *ringBuffer) tryAdd(addr string) bool {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	if _, ok := rb.seen[addr]; ok {
		return false
	}
	if old := rb.ring[rb.pos]; old != "" {
		delete(rb.seen, old)
	}
	rb.ring[rb.pos] = addr
	rb.seen[addr] = struct{}{}
	rb.pos = (rb.pos + 1) % ringSize
	return true
}

// ─── Listener ─────────────────────────────────────────────────────────────────

type Listener struct {
	wsURL  string
	redis  *redisclient.Client
	logger *zap.Logger
	ring   *ringBuffer
}

func New(wsURL string, redis *redisclient.Client, logger *zap.Logger) *Listener {
	return &Listener{
		wsURL:  wsURL,
		redis:  redis,
		logger: logger,
		ring:   &ringBuffer{seen: make(map[string]struct{}, ringSize)},
	}
}

// Run runs the listener with exponential-backoff reconnection.
func (l *Listener) Run(ctx context.Context) {
	backoff := time.Second
	maxBackoff := 30 * time.Second

	for {
		select {
		case <-ctx.Done():
			l.logger.Info("listener shutting down")
			return
		default:
		}

		start := time.Now()
		if err := l.subscribe(ctx); err != nil {
			if ctx.Err() != nil {
				return
			}
			l.logger.Error("listener error – reconnecting",
				zap.Error(err),
				zap.Duration("backoff", backoff),
			)
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
			}
			if time.Since(start) > 30*time.Second {
				backoff = time.Second
			} else {
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
			}
		} else {
			backoff = time.Second
		}
	}
}

func (l *Listener) subscribe(ctx context.Context) error {
	l.logger.Info("connecting to BSC WebSocket", zap.String("url", maskURL(l.wsURL)))

	client, err := ethclient.DialContext(ctx, l.wsURL)
	if err != nil {
		return err
	}
	defer client.Close()

	v2ABI, err := abi.JSON(strings.NewReader(pairCreatedABIJSON))
	if err != nil {
		return err
	}
	v3ABI, err := abi.JSON(strings.NewReader(poolCreatedABIJSON))
	if err != nil {
		return err
	}
	stableABI, err := abi.JSON(strings.NewReader(newStablePairABIJSON))
	if err != nil {
		return err
	}

	v2ID := v2ABI.Events["PairCreated"].ID
	v3ID := v3ABI.Events["PoolCreated"].ID
	stableID := stableABI.Events["NewStableSwapPair"].ID

	query := ethereum.FilterQuery{
		Addresses: []common.Address{
			common.HexToAddress(FactoryV2),
			common.HexToAddress(FactoryV3),
			common.HexToAddress(FactoryStable),
		},
	}

	logs := make(chan types.Log, 1024)
	sub, err := client.SubscribeFilterLogs(ctx, query, logs)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	l.logger.Info("subscribed to PancakeSwap V2+V3+StableSwap factories",
		zap.Int("quote_tokens", len(quoteTokens)),
		zap.String("v3_fee_tiers", "all (0.05%, 0.25%, 1%)"),
	)

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-sub.Err():
			return err
		case log := <-logs:
			if len(log.Topics) == 0 {
				continue
			}
			topic := log.Topics[0]
			switch {
			case topic == v2ID:
				l.handleV2(ctx, &v2ABI, &log)
			case topic == v3ID:
				l.handleV3(ctx, &v3ABI, &log)
			case topic == stableID:
				l.handleStable(ctx, &stableABI, &log)
			}
		}
	}
}

// ─── V2 handler (PairCreated) ─────────────────────────────────────────────────

func (l *Listener) handleV2(ctx context.Context, factoryABI *abi.ABI, log *types.Log) {
	if len(log.Topics) < 3 {
		return
	}
	token0 := common.HexToAddress(log.Topics[1].Hex()).Hex()
	token1 := common.HexToAddress(log.Topics[2].Hex()).Hex()

	type nonIndexed struct {
		Pair           common.Address
		AllPairsLength *big.Int
	}
	var decoded nonIndexed
	if err := factoryABI.UnpackIntoInterface(&decoded, "PairCreated", log.Data); err != nil {
		l.logger.Warn("v2 unpack failed", zap.Error(err))
		return
	}

	meme, quote, sym, ok := extractMemeAndQuote(token0, token1)
	if !ok {
		return
	}
	l.publish(ctx, token0, token1, decoded.Pair.Hex(), meme, quote, sym, "v2", 0, log.BlockNumber)
}

// ─── V3 handler (PoolCreated) ─────────────────────────────────────────────────

func (l *Listener) handleV3(ctx context.Context, poolABI *abi.ABI, log *types.Log) {
	if len(log.Topics) < 3 {
		return
	}
	token0 := common.HexToAddress(log.Topics[1].Hex()).Hex()
	token1 := common.HexToAddress(log.Topics[2].Hex()).Hex()

	type nonIndexed struct {
		TickSpacing *big.Int
		Pool        common.Address
	}
	var decoded nonIndexed
	poolAddr := ""
	if err := poolABI.UnpackIntoInterface(&decoded, "PoolCreated", log.Data); err == nil {
		poolAddr = decoded.Pool.Hex()
	} else if len(log.Data) >= 64 {
		poolAddr = common.BytesToAddress(log.Data[44:64]).Hex()
	}
	if poolAddr == "" {
		return
	}

	meme, quote, sym, ok := extractMemeAndQuote(token0, token1)
	if !ok {
		return
	}

	// Fee tier is the third indexed topic on V3 PoolCreated events.
	var feeTier uint32
	if len(log.Topics) >= 4 {
		feeTier = uint32(new(big.Int).SetBytes(log.Topics[3].Bytes()).Uint64())
	}

	l.publish(ctx, token0, token1, poolAddr, meme, quote, sym, "v3", feeTier, log.BlockNumber)
}

// ─── StableSwap handler ───────────────────────────────────────────────────────

func (l *Listener) handleStable(ctx context.Context, stableABI *abi.ABI, log *types.Log) {
	if len(log.Topics) < 3 {
		return
	}
	token0 := common.HexToAddress(log.Topics[1].Hex()).Hex()
	token1 := common.HexToAddress(log.Topics[2].Hex()).Hex()

	type nonIndexed struct {
		PairContract common.Address
		A            *big.Int
	}
	var decoded nonIndexed
	pairAddr := ""
	if err := stableABI.UnpackIntoInterface(&decoded, "NewStableSwapPair", log.Data); err == nil {
		pairAddr = decoded.PairContract.Hex()
	} else if len(log.Data) >= 32 {
		pairAddr = common.BytesToAddress(log.Data[12:32]).Hex()
	}
	if pairAddr == "" || pairAddr == (common.Address{}).Hex() {
		return
	}

	meme, quote, sym, ok := extractMemeAndQuote(token0, token1)
	if !ok {
		return
	}
	l.publish(ctx, token0, token1, pairAddr, meme, quote, sym, "stable", 0, log.BlockNumber)
}

// ─── Shared publish ───────────────────────────────────────────────────────────

func (l *Listener) publish(ctx context.Context, token0, token1, pairAddr, meme, quote, quoteSymbol, source string, feeTier uint32, block uint64) {
	if !l.ring.tryAdd(strings.ToLower(pairAddr)) {
		l.logger.Debug("duplicate pair skipped", zap.String("pair", pairAddr))
		return
	}

	event := &NewPairEvent{
		Token0:      token0,
		Token1:      token1,
		PairAddress: pairAddr,
		MemeToken:   meme,
		QuoteToken:  quote,
		QuoteSymbol: quoteSymbol,
		Source:      source,
		FeeTier:     feeTier,
		BlockNumber: block,
		Timestamp:   time.Now().Unix(),
	}

	fields := []zap.Field{
		zap.String("source", source),
		zap.String("meme", meme),
		zap.String("quote", quoteSymbol),
		zap.String("pair", pairAddr),
		zap.Uint64("block", block),
	}
	if feeTier > 0 {
		fields = append(fields, zap.Uint32("fee_tier", feeTier))
	}
	l.logger.Info("PairCreated", fields...)

	data, _ := json.Marshal(event)
	if err := l.redis.PublishToStream(ctx, redisclient.StreamNewPairs, json.RawMessage(data)); err != nil {
		l.logger.Error("publish to stream failed", zap.Error(err))
	}
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// extractMemeAndQuote returns (memeToken, quoteToken, quoteSymbol, ok).
// Exactly one of {token0, token1} must be in the quote whitelist.
func extractMemeAndQuote(token0, token1 string) (meme, quote, sym string, ok bool) {
	t0 := strings.ToLower(token0)
	t1 := strings.ToLower(token1)
	sym0, t0IsQuote := quoteTokens[t0]
	sym1, t1IsQuote := quoteTokens[t1]

	switch {
	case t0IsQuote && t1IsQuote:
		return "", "", "", false // both quote — skip (e.g. USDT/WBNB pool)
	case t0IsQuote:
		return token1, token0, sym0, true
	case t1IsQuote:
		return token0, token1, sym1, true
	default:
		return "", "", "", false // neither is a recognised quote token
	}
}

func maskURL(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		last := parts[len(parts)-1]
		if len(last) > 8 {
			parts[len(parts)-1] = last[:4] + "****" + last[len(last)-4:]
		}
	}
	return strings.Join(parts, "/")
}
