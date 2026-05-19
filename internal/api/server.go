// Package api wires HTTP routing, middleware, and handler dependencies.
package api

import (
	"log/slog"
	"net/http"

	"github.com/chessgoddess/chesslens/internal/auth"
	"github.com/chessgoddess/chesslens/internal/config"
	"github.com/chessgoddess/chesslens/internal/db"
	"github.com/chessgoddess/chesslens/internal/middleware"
	"github.com/chessgoddess/chesslens/internal/repository"
	"github.com/chessgoddess/chesslens/internal/service"
	"github.com/chessgoddess/chesslens/internal/worker"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Server struct {
	router      *gin.Engine
	config      *config.Config
	database    *db.DB
	auth        *auth.Service
	queue       *worker.Queue
	bgWorker    *worker.Worker
	aiSvc       *service.AIService
	redisClient *redis.Client
}

func New(cfg *config.Config, database *db.DB, authService *auth.Service) *Server {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Use gin.New() to avoid double middleware (gin.Default adds Logger+Recovery automatically).
	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS(cfg.AllowedOrigins))
	r.Use(middleware.RateLimiter(10, 30)) // 10 req/sec per IP, burst 30

	// Repositories
	userRepo := repository.NewUserRepository(database.Pool)
	gameRepo := repository.NewGameRepository(database.Pool)
	sessionRepo := repository.NewAnalysisSessionRepository(database.Pool)
	moveRepo := repository.NewMoveRepository(database.Pool)
	snapshotRepo := repository.NewSnapshotRepository(database.Pool)
	aiRepo := repository.NewAIExplanationRepository(database.Pool)

	// Analysis service
	analysisSvc, err := service.NewAnalysisService(cfg.StockfishPath, moveRepo, sessionRepo)
	if err != nil {
		slog.Warn("stockfish unavailable", "error", err)
	}

	// Redis
	redisOpts, _ := redis.ParseURL(cfg.RedisURL)
	redisClient := redis.NewClient(redisOpts)

	// AI service
	var aiService *service.AIService
	if cfg.OpenRouterKey != "" {
		aiService = service.NewAIService(cfg.OpenRouterKey, "", redisClient, aiRepo)
	}

	// Queue + background worker
	q, err := worker.NewQueue(cfg.RedisURL)
	if err != nil {
		slog.Warn("redis queue unavailable", "error", err)
	}

	bgWorker := worker.New("main", q, analysisSvc, sessionRepo, gameRepo, snapshotRepo, 2)
	bgWorker.Start()

	// Vision client
	var visionClient *service.VisionClient
	if cfg.OpenRouterKey != "" {
		visionClient = service.NewVisionClient(cfg.OpenRouterKey, "")
	}

	// Handlers
	authHandlers := NewAuthHandlers(authService, userRepo)
	gameHandlers := NewGameHandlers(gameRepo, sessionRepo, analysisSvc, q)
	snapshotHandlers := NewSnapshotHandlers(snapshotRepo)
	aiHandlers := NewAIHandlers(aiService, moveRepo)
	visionHandlers := NewVisionHandlers(visionClient)

	// Routes
	r.GET("/health", healthHandler)
	r.GET("/ready", readyHandler(database, redisClient))

	api := r.Group("/api/v1")
	{
		api.GET("/snapshots/:token", snapshotHandlers.GetByShareToken)

		authGroup := api.Group("/auth")
		{
			authGroup.GET("/google/url", authHandlers.GetGoogleAuthURL)
			authGroup.GET("/google/callback", authHandlers.GoogleCallback)
		}

		protected := api.Group("")
		protected.Use(middleware.Auth(cfg.JWTSecret))
		{
			protected.POST("/games/upload", gameHandlers.UploadGame)
			protected.POST("/analysis", gameHandlers.CreateAnalysis)
			protected.GET("/analysis/:id", gameHandlers.GetAnalysis)
			protected.GET("/games", gameHandlers.ListGames)
			protected.GET("/games/:id", gameHandlers.GetGame)
			protected.DELETE("/games/:id", gameHandlers.DeleteGame)
			protected.POST("/snapshots", gameHandlers.CreateSnapshot)
			protected.GET("/snapshots", snapshotHandlers.ListByUser)

			aiGroup := protected.Group("/ai")
			{
				aiGroup.POST("/explain", aiHandlers.ExplainMove)
				aiGroup.POST("/explain-blunder", aiHandlers.ExplainBlunder)
				aiGroup.GET("/explanation/:move_id", aiHandlers.GetExplanation)
			}

			visionGroup := protected.Group("/vision")
			{
				visionGroup.POST("/image-to-fen", visionHandlers.ImageToFEN)
				visionGroup.POST("/image-to-fen-url", visionHandlers.ImageToFENURL)
			}
		}
	}

	_ = userRepo // used by authHandlers

	return &Server{
		router:      r,
		config:      cfg,
		database:    database,
		auth:        authService,
		queue:       q,
		bgWorker:    bgWorker,
		aiSvc:       aiService,
		redisClient: redisClient,
	}
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) Stop() {
	if s.bgWorker != nil {
		s.bgWorker.Stop()
	}
	if s.queue != nil {
		s.queue.Close()
	}
	if s.redisClient != nil {
		s.redisClient.Close()
	}
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "chesslens"})
}

func readyHandler(database *db.DB, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		if err := database.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready", "reason": "database unavailable"})
			return
		}

		if err := redisClient.Ping(ctx).Err(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready", "reason": "redis unavailable"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	}
}
