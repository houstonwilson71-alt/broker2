package listener

import (
	"context"
	"encoding/json"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	redisclient "github.com/bsc-sniper/backend/internal/redis"
	"go.uber.org/zap"
)

// PancakeSwap V2 Factory on BSC mainnet
const FactoryAddress = "0xcA143Ce32Fe78f1f7019d7d551a6402fC5350c73"

// PairCreated event signature: PairCreated(address,address,address,uint256)
const PairCreatedABI = `[{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"token0","type":"address"},{"indexed":true,"internalType":"address","name":"token1","type":"address"},{"indexed":false,"internalType":"address","name":"pair","type":"address"},{"indexed":false,"internalType":"uint256","name":"allPairsLength","type":"uint256"}],"name":"PairCreated","type":"event"}]`

// WBNB address on BSC
const WBNBAddress = "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095b"

type NewPairEvent struct {
	Token0      string `json:"token0"`
	Token1      string `json:"token1"`
	PairAddress string `json:"pair_address"`
	BlockNumber uint64 `json:"block_number"`
	Timestamp   int64  `json:"timestamp"`
}

type Listener struct {
	wsURL  string
	redis  *redisclient.Client
	logger *zap.Logger
}

func New(wsURL string, redis *redisclient.Client, logger *zap.Logger) *Listener {
	return &Listener{wsURL: wsURL, redis: redis, logger: logger}
}

func (l *Listener) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			l.logger.Info("listener shutting down")
			return
		default:
		}

		if err := l.subscribe(ctx); err != nil {
			l.logger.Error("listener error, reconnecting in 5s", zap.Error(err))
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
			}
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

	factoryABI, err := abi.JSON(strings.NewReader(PairCreatedABI))
	if err != nil {
		return err
	}

	factoryAddr := common.HexToAddress(FactoryAddress)
	query := ethereum.FilterQuery{
		Addresses: []common.Address{factoryAddr},
	}

	logs := make(chan types.Log, 512)
	sub, err := client.SubscribeFilterLogs(ctx, query, logs)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	l.logger.Info("subscribed to PancakeSwap PairCreated events")

	pairCreatedID := factoryABI.Events["PairCreated"].ID

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-sub.Err():
			return err
		case log := <-logs:
			if len(log.Topics) == 0 || log.Topics[0] != pairCreatedID {
				continue
			}
			l.handleLog(ctx, &factoryABI, &log)
		}
	}
}

func (l *Listener) handleLog(ctx context.Context, factoryABI *abi.ABI, log *types.Log) {
	// Topics: [0]=eventID, [1]=token0 (indexed), [2]=token1 (indexed)
	if len(log.Topics) < 3 {
		return
	}

	token0 := common.HexToAddress(log.Topics[1].Hex()).Hex()
	token1 := common.HexToAddress(log.Topics[2].Hex()).Hex()

	// Decode non-indexed data: pair address and allPairsLength
	type nonIndexed struct {
		Pair          common.Address
		AllPairsLength *big.Int
	}
	var decoded nonIndexed
	if err := factoryABI.UnpackIntoInterface(&decoded, "PairCreated", log.Data); err != nil {
		l.logger.Warn("failed to unpack PairCreated data", zap.Error(err))
		return
	}

	pairAddr := decoded.Pair.Hex()

	// Only process pairs that include WBNB (to focus on meme coins vs BNB)
	wbnb := strings.ToLower(WBNBAddress)
	if strings.ToLower(token0) != wbnb && strings.ToLower(token1) != wbnb {
		return
	}

	// Determine which token is the meme coin (non-WBNB side)
	memeToken := token0
	if strings.ToLower(token0) == wbnb {
		memeToken = token1
	}

	event := &NewPairEvent{
		Token0:      token0,
		Token1:      token1,
		PairAddress: pairAddr,
		BlockNumber: log.BlockNumber,
		Timestamp:   time.Now().Unix(),
	}

	l.logger.Info("PairCreated",
		zap.String("meme_token", memeToken),
		zap.String("pair", pairAddr),
		zap.Uint64("block", log.BlockNumber),
	)

	data, _ := json.Marshal(event)
	if err := l.redis.PublishToStream(ctx, redisclient.StreamNewPairs, json.RawMessage(data)); err != nil {
		l.logger.Error("failed to publish to stream", zap.Error(err))
	}
}

// maskURL masks API keys in URLs for safe logging.
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
