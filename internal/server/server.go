package server

import (
	"log"

	"github.com/chessgoddess/chesslens/internal/analysis"
	"github.com/chessgoddess/chesslens/internal/auth"
	"github.com/chessgoddess/chesslens/internal/config"
	"github.com/chessgoddess/chesslens/internal/database"
	"github.com/chessgoddess/chesslens/internal/middleware"
	"github.com/chessgoddess/chesslens/internal/queue"
	"github.com/chessgoddess/chesslens/internal/repository"
	"github.com/chessgoddess/chesslens/internal/worker"
	"github.com/gin-gonic/gin"
)

type Server struct {
	router *gin.Engine
	config *config.Config
	db     *database.Database
	auth   *auth.Service
	q      *queue.Queue
	w      *worker.Worker
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

	// Initialize analysis service
	analysisService, err := analysis.NewService(cfg.StockfishPath, moveRepo, sessionRepo)
	if err != nil {
		log.Printf("Warning: Failed to initialize Stockfish engine: %v", err)
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
		}
	}

	return &Server{
		router: r,
		config: cfg,
		db:     db,
		auth:   authService,
		q:      q,
		w:      w,
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
}

func healthHandler(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok", "service": "chesslens"})
}
