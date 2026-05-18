package server

import (
	"log"

	"github.com/chessgoddess/chesslens/internal/ai"
	"github.com/chessgoddess/chesslens/internal/analysis"
	"github.com/chessgoddess/chesslens/internal/auth"
	"github.com/chessgoddess/chesslens/internal/config"
	"github.com/chessgoddess/chesslens/internal/database"
	"github.com/chessgoddess/chesslens/internal/middleware"
	"github.com/chessgoddess/chesslens/internal/queue"
	"github.com/chessgoddess/chesslens/internal/repository"
	"github.com/chessgoddess/chesslens/internal/vision"
	"github.com/chessgoddess/chesslens/internal/worker"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Server struct {
	router *gin.Engine
	config *config.Config
	db     *database.Database
	auth   *auth.Service
	q      *queue.Queue
	w      *worker.Worker
	aiSvc  *ai.Service
	redis  *redis.Client
}

func New(cfg *config.Config, db *database.Database, authService *auth.Service) *Server {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Global middleware
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())

	// Health check
	r.GET("/health", healthHandler)

	// Initialize repositories
	gameRepo := repository.NewGameRepository(db.Pool)
	sessionRepo := repository.NewAnalysisSessionRepository(db.Pool)
	moveRepo := repository.NewMoveRepository(db.Pool)
	snapshotRepo := repository.NewSnapshotRepository(db.Pool)
	aiRepo := repository.NewAIExplanationRepository(db.Pool)

	// Initialize analysis service
	analysisService, err := analysis.NewService(cfg.StockfishPath, moveRepo, sessionRepo)
	if err != nil {
		log.Printf("Warning: Failed to initialize Stockfish engine: %v", err)
	}

	// Initialize Redis
	redisOpts, _ := redis.ParseURL(cfg.RedisURL)
	redisClient := redis.NewClient(redisOpts)

	// Initialize AI service
	var aiService *ai.Service
	if cfg.OpenRouterKey != "" {
		aiService = ai.NewService(cfg.OpenRouterKey, "", redisClient, aiRepo)
	}

	// Initialize Redis queue
	q, err := queue.New(cfg.RedisURL)
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis queue: %v", err)
	}

	// Initialize worker
	w := worker.New("main", q, analysisService, sessionRepo, gameRepo, snapshotRepo, 2)
	w.Start()

	// Initialize handlers
	authHandlers := NewAuthHandlers(authService)
	gameHandlers := NewGameHandlers(gameRepo, sessionRepo, analysisService, q)
	snapshotHandlers := NewSnapshotHandlers(snapshotRepo)
	aiHandlers := NewAIHandlers(aiService, moveRepo)

	// Initialize vision client
	var visionClient *vision.ImageToFENClient
	if cfg.OpenRouterKey != "" {
		visionClient = vision.NewImageToFENClient(cfg.OpenRouterKey, "")
	}
	visionHandlers := NewVisionHandlers(visionClient)

	// API routes
	api := r.Group("/api/v1")
	{
		// Public routes
		api.GET("/snapshots/:token", snapshotHandlers.GetByShareToken)

		// Auth routes
		auth := api.Group("/auth")
		{
			auth.GET("/google/url", authHandlers.GetGoogleAuthURL)
			auth.GET("/google/callback", authHandlers.GoogleCallback)
		}

		// Protected routes
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
			
			// AI routes
			ai := protected.Group("/ai")
			{
				ai.POST("/explain", aiHandlers.ExplainMove)
				ai.POST("/explain-blunder", aiHandlers.ExplainBlunder)
				ai.GET("/explanation/:move_id", aiHandlers.GetExplanation)
			}

			// Vision routes
			vision := protected.Group("/vision")
			{
				vision.POST("/image-to-fen", visionHandlers.ImageToFEN)
				vision.POST("/image-to-fen-url", visionHandlers.ImageToFENURL)
			}
		}
	}

	return &Server{
		router: r,
		config: cfg,
		db:     db,
		auth:   authService,
		q:      q,
		w:      w,
		aiSvc:  aiService,
		redis:  redisClient,
	}
}

func (s *Server) Start(addr string) error {
	return s.router.Run(addr)
}

func (s *Server) Stop() {
	if s.w != nil {
		s.w.Stop()
	}
	if s.q != nil {
		s.q.Close()
	}
	if s.redis != nil {
		s.redis.Close()
	}
}

func healthHandler(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok", "service": "chesslens"})
}
