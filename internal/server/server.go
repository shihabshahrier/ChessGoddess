package server

import (
	"github.com/chessgoddess/chesslens/internal/config"
	"github.com/chessgoddess/chesslens/internal/middleware"
	"github.com/gin-gonic/gin"
)

type Server struct {
	router *gin.Engine
	config *config.Config
}

func New(cfg *config.Config) *Server {
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

	// API routes
	api := r.Group("/api/v1")
	{
		// Public routes
		api.GET("/snapshots/:id", getSnapshotHandler)

		// Auth routes
		auth := api.Group("/auth")
		{
			auth.GET("/google/url", getGoogleAuthURLHandler)
			auth.GET("/google/callback", googleCallbackHandler)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.Auth(cfg.JWTSecret))
		{
			protected.POST("/games/upload", uploadGameHandler)
			protected.POST("/analysis", createAnalysisHandler)
			protected.GET("/analysis/:id", getAnalysisHandler)
			protected.GET("/games", listGamesHandler)
			protected.GET("/games/:id", getGameHandler)
			protected.DELETE("/games/:id", deleteGameHandler)
		}
	}

	return &Server{
		router: r,
		config: cfg,
	}
}

func (s *Server) Start(addr string) error {
	return s.router.Run(addr)
}

func healthHandler(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok", "service": "chesslens"})
}

// Placeholder handlers - will be implemented in respective modules
func getSnapshotHandler(c *gin.Context)        { c.JSON(200, gin.H{"message": "not implemented"}) }
func getGoogleAuthURLHandler(c *gin.Context)   { c.JSON(200, gin.H{"message": "not implemented"}) }
func googleCallbackHandler(c *gin.Context)     { c.JSON(200, gin.H{"message": "not implemented"}) }
func uploadGameHandler(c *gin.Context)         { c.JSON(200, gin.H{"message": "not implemented"}) }
func createAnalysisHandler(c *gin.Context)     { c.JSON(200, gin.H{"message": "not implemented"}) }
func getAnalysisHandler(c *gin.Context)        { c.JSON(200, gin.H{"message": "not implemented"}) }
func listGamesHandler(c *gin.Context)          { c.JSON(200, gin.H{"message": "not implemented"}) }
func getGameHandler(c *gin.Context)            { c.JSON(200, gin.H{"message": "not implemented"}) }
func deleteGameHandler(c *gin.Context)         { c.JSON(200, gin.H{"message": "not implemented"}) }
