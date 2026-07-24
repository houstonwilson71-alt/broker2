package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/bsc-sniper/backend/internal/api"
	"github.com/bsc-sniper/backend/internal/config"
	"github.com/bsc-sniper/backend/internal/db"
	"github.com/bsc-sniper/backend/internal/executor"
	"github.com/bsc-sniper/backend/internal/filter"
	"github.com/bsc-sniper/backend/internal/listener"
	"github.com/bsc-sniper/backend/internal/monitor"
	redisclient "github.com/bsc-sniper/backend/internal/redis"
	"go.uber.org/zap"
)

const (
	filterWorkers   = 4
	executorWorkers = 4
)

// botState constants
const (
	stateIdle     = "idle"
	stateScanning = "scanning"
)

// Bot orchestrates all components.
type Bot struct {
	cfg      *config.Config
	rpcWS    *ethclient.Client
	rpcHTTP  *ethclient.Client
	redis    *redisclient.Client
	database *db.DB
	listener *listener.Listener
	filter   *filter.Engine
	executor *executor.Executor
	monitor  *monitor.Monitor
	logger   *zap.Logger

	mu      sync.Mutex
	running bool
	cancel  context.CancelFunc
	wg      sync.WaitGroup

	stateStr atomic.Value // string: "idle" | "scanning" | "buying" | "selling"
}

func NewBot(cfg *config.Config, rpcWS, rpcHTTP *ethclient.Client, redis *redisclient.Client, database *db.DB, logger *zap.Logger) (*Bot, error) {
	lst := listener.New(cfg.BSCRPCWebSocket, redis, logger.Named("listener"))

	filterEngine, err := filter.New(cfg, rpcHTTP, redis, database, logger.Named("filter"))
	if err != nil {
		return nil, fmt.Errorf("filter engine: %w", err)
	}

	exec, err := executor.New(cfg, rpcHTTP, redis, database, logger.Named("executor"))
	if err != nil {
		return nil, fmt.Errorf("executor: %w", err)
	}

	mon, err := monitor.New(cfg, rpcHTTP, redis, database, exec, logger.Named("monitor"))
	if err != nil {
		return nil, fmt.Errorf("monitor: %w", err)
	}

	b := &Bot{
		cfg:      cfg,
		rpcWS:    rpcWS,
		rpcHTTP:  rpcHTTP,
		redis:    redis,
		database: database,
		listener: lst,
		filter:   filterEngine,
		executor: exec,
		monitor:  mon,
		logger:   logger,
	}
	b.stateStr.Store(stateIdle)
	return b, nil
}

func (b *Bot) State() string {
	s, _ := b.stateStr.Load().(string)
	if s == "" {
		return stateIdle
	}
	// If executor has active buys, report "buying"
	if b.executor != nil && b.executor.BuyingActive.Load() > 0 {
		return "buying"
	}
	if b.filter != nil && b.filter.Active.Load() > 0 {
		return "filtering"
	}
	return s
}

func (b *Bot) Start() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.running {
		return fmt.Errorf("already running")
	}

	ctx, cancel := context.WithCancel(context.Background())
	b.cancel = cancel
	b.running = true
	b.stateStr.Store(stateScanning)

	if err := b.monitor.LoadFromDB(ctx); err != nil {
		b.logger.Warn("load positions from db", zap.Error(err))
	}
	if err := b.database.SetBotRunning(ctx, true); err != nil {
		b.logger.Warn("db set running", zap.Error(err))
	}

	// Listener: 1 goroutine
	b.wg.Add(1)
	go func() { defer b.wg.Done(); b.listener.Run(ctx) }()

	// Filter: 4 parallel workers
	for i := 0; i < filterWorkers; i++ {
		id := i
		b.wg.Add(1)
		go func() { defer b.wg.Done(); b.filter.Run(ctx, id) }()
	}

	// Executor: 4 parallel workers
	for i := 0; i < executorWorkers; i++ {
		id := i
		b.wg.Add(1)
		go func() { defer b.wg.Done(); b.executor.Run(ctx, id) }()
	}

	// Monitor: 1 goroutine (not in wg since it handles its own shutdown via ctx)
	go b.monitor.Run(ctx)

	b.logger.Info("bot started",
		zap.Bool("live_trading", b.cfg.LiveTradingEnabled),
		zap.Float64("buy_bnb", b.cfg.BuyAmountBNB),
		zap.Int("filter_workers", filterWorkers),
		zap.Int("executor_workers", executorWorkers),
	)
	return nil
}

func (b *Bot) Stop() error {
	b.mu.Lock()
	if !b.running {
		b.mu.Unlock()
		return fmt.Errorf("not running")
	}
	b.logger.Warn("bot stopping — force selling all open positions first")

	// Force sell every open position before shutting down components.
	sellCtx, sellCancel := context.WithTimeout(context.Background(), 5*time.Minute)
	b.monitor.SellAllPositions(sellCtx)
	sellCancel()

	b.cancel()
	b.wg.Wait()
	b.running = false
	b.stateStr.Store(stateIdle)
	b.mu.Unlock()

	// Print and persist the unsold token board after force sells have attempted.
	b.printUnsoldTokenBoard()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := b.database.SetBotRunning(ctx, false); err != nil {
		b.logger.Warn("db set stopped", zap.Error(err))
	}

	b.logger.Info("bot stopped")
	return nil
}

// printUnsoldTokenBoard queries positions that are still bought or unsellable,
// prints a console table, and writes contract addresses to unsold_tokens.txt.
func (b *Bot) printUnsoldTokenBoard() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	positions, err := b.database.ListPositionsByStatuses(ctx, []string{"bought", "unsellable"})
	if err != nil {
		b.logger.Error("query unsold positions", zap.Error(err))
		return
	}
	if len(positions) == 0 {
		b.logger.Info("unsold token board: no leftover positions")
		return
	}

	var lines []string
	lines = append(lines, "\n═══════════ UNSOLD TOKEN BOARD ═══════════")
	lines = append(lines, fmt.Sprintf("%-12s | %-42s | %-66s | %-24s | %-12s", "Symbol", "Contract", "Buy Tx", "Amount", "Status"))
	lines = append(lines, strings.Repeat("-", 165))
	for _, p := range positions {
		buyTx := "unknown"
		trades, err := b.database.ListTradesByToken(ctx, p.TokenAddress)
		if err == nil {
			for _, t := range trades {
				if t.Side == "buy" && t.Status == "confirmed" {
					buyTx = t.TxHash
					break
				}
			}
		}
		lines = append(lines, fmt.Sprintf("%-12s | %-42s | %-66s | %-24s | %-12s",
			p.TokenSymbol, p.TokenAddress, buyTx, p.AmountTokens, p.Status))
	}
	lines = append(lines, "═══════════════════════════════════════════")
	board := strings.Join(lines, "\n")
	fmt.Println(board)
	b.logger.Info(board)

	// Write contract addresses to unsold_tokens.txt, one per line.
	var addrs []string
	for _, p := range positions {
		addrs = append(addrs, p.TokenAddress)
	}
	if err := os.WriteFile("unsold_tokens.txt", []byte(strings.Join(addrs, "\n")+"\n"), 0644); err != nil {
		b.logger.Error("write unsold_tokens.txt", zap.Error(err))
	} else {
		b.logger.Info("wrote unsold_tokens.txt", zap.Int("count", len(addrs)))
	}
}

// EmergencyStop sells all open positions immediately then stops the bot.
func (b *Bot) EmergencyStop(ctx context.Context) error {
	b.logger.Warn("⚠️  EMERGENCY STOP — selling all positions")

	sellCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	positions, err := b.database.ListPositionsByStatuses(sellCtx, []string{"bought", "partial"})
	if err != nil {
		b.logger.Error("list positions for emergency sell", zap.Error(err))
	} else {
		// Sell sequentially to avoid nonce collisions from parallel tx submission.
		for _, pos := range positions {
			if err := b.executor.ExecuteSell(sellCtx, pos, 100, "force"); err != nil {
				b.logger.Error("emergency sell failed",
					zap.String("token", pos.TokenAddress),
					zap.String("symbol", pos.TokenSymbol),
					zap.Error(err),
				)
			} else {
				b.logger.Info("emergency sell completed",
					zap.String("token", pos.TokenAddress),
					zap.String("symbol", pos.TokenSymbol),
				)
			}
		}
	}

	if b.IsRunning() {
		return b.Stop()
	}
	return nil
}

func (b *Bot) IsRunning() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.running
}

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("BSC Sniper starting up – hardened multi-version build")

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("load config", zap.Error(err))
	}

	// WebSocket RPC
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	rpcWS, err := ethclient.DialContext(ctx, cfg.BSCRPCWebSocket)
	cancel()
	if err != nil {
		logger.Fatal("connect BSC WebSocket", zap.Error(err))
	}
	defer rpcWS.Close()

	// HTTP RPC
	ctx, cancel = context.WithTimeout(context.Background(), 15*time.Second)
	rpcHTTP, err := ethclient.DialContext(ctx, cfg.BSCRPCHTTP)
	cancel()
	if err != nil {
		logger.Fatal("connect BSC HTTP", zap.Error(err))
	}
	defer rpcHTTP.Close()

	// Redis
	redis, err := redisclient.New(cfg.RedisURL, logger.Named("redis"))
	if err != nil {
		logger.Fatal("connect redis", zap.Error(err))
	}
	defer redis.Close()

	// Postgres
	ctx, cancel = context.WithTimeout(context.Background(), 15*time.Second)
	database, err := db.New(ctx, cfg.DatabaseURL, logger.Named("db"))
	cancel()
	if err != nil {
		logger.Fatal("connect database", zap.Error(err))
	}
	defer database.Close()

	// Bot
	bot, err := NewBot(cfg, rpcWS, rpcHTTP, redis, database, logger.Named("bot"))
	if err != nil {
		logger.Fatal("create bot", zap.Error(err))
	}

	// API server
	server := api.NewServer(cfg, database, redis, bot, logger.Named("api"))
	router := server.Router()

	httpSrv := &http.Server{
		Addr:         ":" + cfg.APIPort,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		logger.Info("API server listening", zap.String("addr", httpSrv.Addr))
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("http server", zap.Error(err))
		}
	}()

	// OS signal handling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down gracefully")

	if bot.IsRunning() {
		_ = bot.Stop()
	}

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutCancel()
	if err := httpSrv.Shutdown(shutCtx); err != nil {
		logger.Error("http shutdown", zap.Error(err))
	}
	logger.Info("bye")
}
