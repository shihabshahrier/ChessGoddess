// Package api wires HTTP routing, middleware, and handler dependencies.
package api

import (
	"log/slog"
	"net/http"

	"github.com/chessgoddess/chessgoddess/internal/auth"
	"github.com/chessgoddess/chessgoddess/internal/config"
	"github.com/chessgoddess/chessgoddess/internal/db"
	"github.com/chessgoddess/chessgoddess/internal/engine"
	"github.com/chessgoddess/chessgoddess/internal/middleware"
	"github.com/chessgoddess/chessgoddess/internal/repository"
	"github.com/chessgoddess/chessgoddess/internal/service"
	"github.com/chessgoddess/chessgoddess/internal/worker"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Server struct {
	router      *gin.Engine
	config      *config.Config
	database    *db.DB
	auth        *auth.Service
	queue       worker.JobQueue
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
	analysisSvc, err := service.NewAnalysisService(cfg.StockfishPath, cfg.EnginePoolSize, moveRepo, sessionRepo)
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

	// Queue — SQS in production, Redis locally
	var q worker.JobQueue
	switch cfg.QueueProvider {
	case "sqs":
		q, err = worker.NewSQSQueue(cfg.SQSAnalysisURL, cfg.SQSSnapshotURL, cfg.SQSAIExplainURL)
		if err != nil {
			slog.Warn("sqs queue unavailable", "error", err)
		}
	default:
		q, err = worker.NewRedisQueue(cfg.RedisURL)
		if err != nil {
			slog.Warn("redis queue unavailable", "error", err)
		}
	}

	// Background worker
	var bgWorker *worker.Worker
	if cfg.WorkerEnabled {
		bgWorker = worker.New("main", q, analysisSvc, sessionRepo, gameRepo, snapshotRepo, 2)
		bgWorker.Start()
	}

	// Vision client
	var visionClient *service.VisionClient
	if cfg.OpenRouterKey != "" {
		visionClient = service.NewVisionClient(cfg.OpenRouterKey, "")
	}

	// Engine pool — shared with the analysis service when Stockfish is available.
	var enginePool *engine.Pool
	if analysisSvc != nil {
		enginePool = analysisSvc.Pool()
	}

	// Handlers
	authHandlers := NewAuthHandlers(authService, userRepo, cfg)
	gameHandlers := NewGameHandlers(gameRepo, sessionRepo, analysisSvc, q)
	snapshotHandlers := NewSnapshotHandlers(snapshotRepo)
	aiHandlers := NewAIHandlers(aiService, moveRepo)
	visionHandlers := NewVisionHandlers(visionClient)
	engineHandlers := NewEngineHandlers(enginePool)

	// Routes
	r.GET("/health", healthHandler)
	r.GET("/ready", readyHandler(database, redisClient))

	api := r.Group("/api/v1")
	{
		api.GET("/snapshots/:token", snapshotHandlers.GetByShareToken)

		// Public — board eval. Bounded by the rate limiter and the engine pool;
		// the heavy continuous eval runs client-side (WASM).
		api.POST("/engine/evaluate", engineHandlers.Evaluate)

		// Public — stateless board-image recognition (no persistence, no user data).
		visionGroup := api.Group("/vision")
		{
			visionGroup.POST("/image-to-fen", visionHandlers.ImageToFEN)
			visionGroup.POST("/image-to-fen-url", visionHandlers.ImageToFENURL)
		}

		authGroup := api.Group("/auth")
		{
			authGroup.GET("/google/url", authHandlers.GetGoogleAuthURL)
			authGroup.GET("/google/callback", authHandlers.GoogleCallback)
			authGroup.POST("/logout", authHandlers.Logout)
		}

		protected := api.Group("")
		protected.Use(middleware.Auth(cfg.JWTSecret))
		{
			protected.GET("/auth/me", authHandlers.Me)

			protected.POST("/games/upload", gameHandlers.UploadGame)
			protected.POST("/analysis", gameHandlers.CreateAnalysis)
			protected.GET("/analysis/:id", gameHandlers.GetAnalysis)
			protected.GET("/analysis/:id/moves", gameHandlers.GetAnalysisMoves)
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
		}
	}

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
	c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "chessgoddess"})
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
