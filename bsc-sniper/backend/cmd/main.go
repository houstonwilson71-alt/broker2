package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
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

	return &Bot{
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
	}, nil
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

	if err := b.monitor.LoadFromDB(ctx); err != nil {
		b.logger.Warn("load positions from db failed", zap.Error(err))
	}

	if err := b.database.SetBotRunning(ctx, true); err != nil {
		b.logger.Warn("db set bot running", zap.Error(err))
	}

	b.wg.Add(3)
	go func() { defer b.wg.Done(); b.listener.Run(ctx) }()
	go func() { defer b.wg.Done(); b.filter.Run(ctx) }()
	go func() { defer b.wg.Done(); b.executor.Run(ctx) }()
	go func() { b.monitor.Run(ctx) }()

	b.logger.Info("bot started",
		zap.Bool("live_trading", b.cfg.LiveTradingEnabled),
		zap.Float64("buy_bnb", b.cfg.BuyAmountBNB),
	)
	return nil
}

func (b *Bot) Stop() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.running {
		return fmt.Errorf("not running")
	}

	b.cancel()
	b.wg.Wait()
	b.running = false

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := b.database.SetBotRunning(ctx, false); err != nil {
		b.logger.Warn("db set bot stopped", zap.Error(err))
	}

	b.logger.Info("bot stopped")
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

	logger.Info("BSC Sniper starting up")

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("load config", zap.Error(err))
	}

	// Connect to BSC via WebSocket
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	rpcWS, err := ethclient.DialContext(ctx, cfg.BSCRPCWebSocket)
	cancel()
	if err != nil {
		logger.Fatal("connect to BSC WebSocket", zap.Error(err))
	}
	defer rpcWS.Close()

	// Connect via HTTP for calls
	ctx, cancel = context.WithTimeout(context.Background(), 15*time.Second)
	rpcHTTP, err := ethclient.DialContext(ctx, cfg.BSCRPCHTTP)
	cancel()
	if err != nil {
		logger.Fatal("connect to BSC HTTP", zap.Error(err))
	}
	defer rpcHTTP.Close()

	// Connect to Redis
	redis, err := redisclient.New(cfg.RedisURL, logger.Named("redis"))
	if err != nil {
		logger.Fatal("connect to redis", zap.Error(err))
	}
	defer redis.Close()

	// Connect to DB
	ctx, cancel = context.WithTimeout(context.Background(), 15*time.Second)
	database, err := db.New(ctx, cfg.DatabaseURL, logger.Named("db"))
	cancel()
	if err != nil {
		logger.Fatal("connect to database", zap.Error(err))
	}
	defer database.Close()

	// Create bot
	bot, err := NewBot(cfg, rpcWS, rpcHTTP, redis, database, logger.Named("bot"))
	if err != nil {
		logger.Fatal("create bot", zap.Error(err))
	}

	// Create API server
	server := api.NewServer(cfg, database, redis, bot, logger.Named("api"))
	router := server.Router()

	httpSrv := &http.Server{
		Addr:         ":" + cfg.APIPort,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start HTTP server
	go func() {
		logger.Info("API server listening", zap.String("addr", httpSrv.Addr))
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("http server", zap.Error(err))
		}
	}()

	// Wait for OS signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down")

	if bot.IsRunning() {
		_ = bot.Stop()
	}

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutCancel()
	if err := httpSrv.Shutdown(shutCtx); err != nil {
		logger.Error("http shutdown", zap.Error(err))
	}

	logger.Info("bye")
}
