package api

import (
	"context"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/bsc-sniper/backend/internal/config"
	"github.com/bsc-sniper/backend/internal/db"
	redisclient "github.com/bsc-sniper/backend/internal/redis"
	"go.uber.org/zap"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 4096,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type BotController interface {
	Start() error
	Stop() error
	IsRunning() bool
	EmergencyStop(ctx context.Context) error
	State() string // "idle" | "scanning" | "buying" | "selling"
}

type Server struct {
	cfg       *config.Config
	db        *db.DB
	redis     *redisclient.Client
	bot       BotController
	logger    *zap.Logger
	wsClients atomic.Int64
}

func NewServer(cfg *config.Config, database *db.DB, redis *redisclient.Client, bot BotController, logger *zap.Logger) *Server {
	return &Server{cfg: cfg, db: database, redis: redis, bot: bot, logger: logger}
}

func (s *Server) Router() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	api := r.Group("/api")
	{
		api.GET("/health", s.health)
		api.POST("/bot/start", s.botStart)
		api.POST("/bot/stop", s.botStop)
		api.POST("/bot/emergency-stop", s.botEmergencyStop)
		api.GET("/bot/status", s.botStatus)
		api.GET("/tokens", s.listTokens)
		api.GET("/trades", s.listTrades)
		api.GET("/positions", s.listPositions)
		api.GET("/config", s.getConfig)
		api.PUT("/config", s.updateConfig)
		api.GET("/ws", s.handleWS)
	}

	return r
}

// GET /api/health
func (s *Server) health(c *gin.Context) {
	ctx := c.Request.Context()
	botState, _ := s.db.GetBotState(ctx)
	resp := gin.H{
		"status":    "ok",
		"timestamp": time.Now().UTC(),
		"state":     s.bot.State(),
	}
	if botState != nil {
		resp["bot_running"] = botState.Running
		resp["pairs_seen"] = botState.PairsSeen
		resp["pairs_passed"] = botState.PairsPassed
		resp["trades_total"] = botState.TradesTotal
	}
	c.JSON(http.StatusOK, resp)
}

// POST /api/bot/start
func (s *Server) botStart(c *gin.Context) {
	if s.bot.IsRunning() {
		c.JSON(http.StatusConflict, gin.H{"error": "bot already running"})
		return
	}
	if err := s.bot.Start(); err != nil {
		s.logger.Error("bot start failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "started", "timestamp": time.Now().UTC()})
}

// POST /api/bot/stop
func (s *Server) botStop(c *gin.Context) {
	if !s.bot.IsRunning() {
		c.JSON(http.StatusConflict, gin.H{"error": "bot not running"})
		return
	}
	if err := s.bot.Stop(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "stopped", "timestamp": time.Now().UTC()})
}

// POST /api/bot/emergency-stop — immediately sells all positions then stops
func (s *Server) botEmergencyStop(c *gin.Context) {
	s.logger.Warn("EMERGENCY STOP requested via API")
	if err := s.bot.EmergencyStop(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":    "emergency_stopped",
		"timestamp": time.Now().UTC(),
	})
}

// GET /api/bot/status
func (s *Server) botStatus(c *gin.Context) {
	ctx := c.Request.Context()
	state, err := s.db.GetBotState(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"running":      state.Running,
		"state":        s.bot.State(),
		"started_at":   state.StartedAt,
		"stopped_at":   state.StoppedAt,
		"pairs_seen":   state.PairsSeen,
		"pairs_passed": state.PairsPassed,
		"trades_total": state.TradesTotal,
		"ws_clients":   s.wsClients.Load(),
	})
}

// GET /api/tokens?limit=50
func (s *Server) listTokens(c *gin.Context) {
	limit := 50
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	tokens, err := s.db.ListTokens(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": tokens, "count": len(tokens)})
}

// GET /api/trades?limit=50
func (s *Server) listTrades(c *gin.Context) {
	limit := 50
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	trades, err := s.db.ListTrades(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": trades, "count": len(trades)})
}

// GET /api/positions?status=open
func (s *Server) listPositions(c *gin.Context) {
	status := c.DefaultQuery("status", "")
	positions, err := s.db.ListPositions(c.Request.Context(), status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": positions, "count": len(positions)})
}

// GET /api/config
func (s *Server) getConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"live_trading_enabled": s.cfg.LiveTradingEnabled,
		"buy_amount_bnb":       s.cfg.BuyAmountBNB,
		"slippage_bps":         s.cfg.SlippageBPS,
		"min_liquidity_usd":    s.cfg.MinLiquidityUSD,
		"max_age_sec":          s.cfg.MaxAgeSec,
		"min_holders":          s.cfg.MinHolders,
		"max_top10_pct":        s.cfg.MaxTop10Pct,
		"max_rug_score":        s.cfg.MaxRugScore,
		"take_profit_1_pct":    s.cfg.TakeProfit1Pct,
		"trailing_stop_pct":    s.cfg.TrailingStopPct,
	})
}

// PUT /api/config
func (s *Server) updateConfig(c *gin.Context) {
	var body struct {
		LiveTradingEnabled *bool    `json:"live_trading_enabled"`
		BuyAmountBNB       *float64 `json:"buy_amount_bnb"`
		SlippageBPS        *int     `json:"slippage_bps"`
		MinLiquidityUSD    *float64 `json:"min_liquidity_usd"`
		MaxAgeSec          *int64   `json:"max_age_sec"`
		MinHolders         *int     `json:"min_holders"`
		MaxTop10Pct        *float64 `json:"max_top10_pct"`
		TakeProfit1Pct     *float64 `json:"take_profit_1_pct"`
		TrailingStopPct    *float64 `json:"trailing_stop_pct"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.LiveTradingEnabled != nil {
		s.cfg.LiveTradingEnabled = *body.LiveTradingEnabled
	}
	if body.BuyAmountBNB != nil && *body.BuyAmountBNB > 0 {
		s.cfg.BuyAmountBNB = *body.BuyAmountBNB
	}
	if body.SlippageBPS != nil && *body.SlippageBPS > 0 {
		s.cfg.SlippageBPS = *body.SlippageBPS
	}
	if body.MinLiquidityUSD != nil {
		s.cfg.MinLiquidityUSD = *body.MinLiquidityUSD
	}
	if body.MaxAgeSec != nil {
		s.cfg.MaxAgeSec = *body.MaxAgeSec
	}
	if body.MinHolders != nil {
		s.cfg.MinHolders = *body.MinHolders
	}
	if body.MaxTop10Pct != nil {
		s.cfg.MaxTop10Pct = *body.MaxTop10Pct
	}
	if body.TakeProfit1Pct != nil {
		s.cfg.TakeProfit1Pct = *body.TakeProfit1Pct
	}
	if body.TrailingStopPct != nil {
		s.cfg.TrailingStopPct = *body.TrailingStopPct
	}
	s.getConfig(c)
}

// GET /api/ws — WebSocket endpoint streaming Redis pub/sub events
func (s *Server) handleWS(c *gin.Context) {
	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		s.logger.Error("ws upgrade failed", zap.Error(err))
		return
	}
	defer conn.Close()

	s.wsClients.Add(1)
	defer s.wsClients.Add(-1)

	ctx := c.Request.Context()
	msgCh, unsub := s.redis.Subscribe(ctx, redisclient.PubSubEvents)
	defer unsub()

	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	for {
		select {
		case msg, ok := <-msgCh:
			if !ok {
				return
			}
			conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
				return
			}
		case <-pingTicker.C:
			conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
