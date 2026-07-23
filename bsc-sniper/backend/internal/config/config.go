package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// RPC
	BSCRPCWebSocket string
	BSCRPCHTTP      string

	// Wallet
	PrivateKey string

	// Trading
	LiveTradingEnabled bool
	BuyAmountBNB       float64
	SlippageBPS        int

	// Filters
	MinLiquidityUSD float64
	MaxAgeSec       int64
	MinHolders      int
	MaxTop10Pct     float64
	MaxRugScore     int

	// Exit strategy
	TakeProfit1Pct   float64
	TrailingStopPct  float64

	// Infrastructure
	BloxrouteURL string
	DatabaseURL  string
	RedisURL     string

	// API
	APIPort string
}

func Load() (*Config, error) {
	// Load .env if present (ignore error — env vars may already be set)
	_ = godotenv.Load()

	cfg := &Config{}

	cfg.BSCRPCWebSocket = requireEnv("BSC_RPC_WS")
	cfg.BSCRPCHTTP = requireEnv("BSC_RPC_HTTP")
	cfg.PrivateKey = requireEnv("PRIVATE_KEY")

	cfg.LiveTradingEnabled = getEnvBool("LIVE_TRADING_ENABLED", false)

	var err error
	cfg.BuyAmountBNB, err = getEnvFloat("BUY_AMOUNT_BNB", 0.0005)
	if err != nil {
		return nil, fmt.Errorf("BUY_AMOUNT_BNB: %w", err)
	}
	cfg.SlippageBPS, err = getEnvInt("SLIPPAGE_BPS", 1500)
	if err != nil {
		return nil, fmt.Errorf("SLIPPAGE_BPS: %w", err)
	}
	cfg.MinLiquidityUSD, err = getEnvFloat("MIN_LIQUIDITY_USD", 1000)
	if err != nil {
		return nil, fmt.Errorf("MIN_LIQUIDITY_USD: %w", err)
	}
	cfg.MaxAgeSec, err = getEnvInt64("MAX_AGE_SEC", 300)
	if err != nil {
		return nil, fmt.Errorf("MAX_AGE_SEC: %w", err)
	}
	cfg.MinHolders, err = getEnvInt("MIN_HOLDERS", 25)
	if err != nil {
		return nil, fmt.Errorf("MIN_HOLDERS: %w", err)
	}
	cfg.MaxTop10Pct, err = getEnvFloat("MAX_TOP10_PCT", 35)
	if err != nil {
		return nil, fmt.Errorf("MAX_TOP10_PCT: %w", err)
	}
	cfg.MaxRugScore, err = getEnvInt("MAX_RUG_SCORE", 2)
	if err != nil {
		return nil, fmt.Errorf("MAX_RUG_SCORE: %w", err)
	}
	cfg.TakeProfit1Pct, err = getEnvFloat("TAKE_PROFIT_1_PCT", 100)
	if err != nil {
		return nil, fmt.Errorf("TAKE_PROFIT_1_PCT: %w", err)
	}
	cfg.TrailingStopPct, err = getEnvFloat("TRAILING_STOP_PCT", 25)
	if err != nil {
		return nil, fmt.Errorf("TRAILING_STOP_PCT: %w", err)
	}

	cfg.BloxrouteURL = os.Getenv("BLOXROUTE_URL")
	cfg.DatabaseURL = getEnvDefault("DATABASE_URL", "postgres://sniper:sniper@postgres:5432/sniper?sslmode=disable")
	cfg.RedisURL = getEnvDefault("REDIS_URL", "redis://redis:6379/0")
	cfg.APIPort = getEnvDefault("API_PORT", "8080")

	return cfg, nil
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required environment variable %q is not set", key))
	}
	return v
}

func getEnvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func getEnvFloat(key string, def float64) (float64, error) {
	v := os.Getenv(key)
	if v == "" {
		return def, nil
	}
	return strconv.ParseFloat(v, 64)
}

func getEnvInt(key string, def int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return def, nil
	}
	i, err := strconv.Atoi(v)
	return i, err
}

func getEnvInt64(key string, def int64) (int64, error) {
	v := os.Getenv(key)
	if v == "" {
		return def, nil
	}
	i, err := strconv.ParseInt(v, 10, 64)
	return i, err
}
