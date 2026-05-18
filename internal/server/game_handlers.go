package server

import (
	"context"

	"github.com/chessgoddess/chesslens/internal/analysis"
	"github.com/chessgoddess/chesslens/internal/game"
	"github.com/chessgoddess/chesslens/internal/models"
	"github.com/chessgoddess/chesslens/internal/repository"
	"github.com/gin-gonic/gin"
)

type GameHandlers struct {
	gameRepo        *repository.GameRepository
	sessionRepo     *repository.AnalysisSessionRepository
	analysisService *analysis.Service
}

func NewGameHandlers(gameRepo *repository.GameRepository, sessionRepo *repository.AnalysisSessionRepository, analysisService *analysis.Service) *GameHandlers {
	return &GameHandlers{
		gameRepo:        gameRepo,
		sessionRepo:     sessionRepo,
		analysisService: analysisService,
	}
}

type UploadGameRequest struct {
	PGN string `json:"pgn" binding:"required"`
}

func (h *GameHandlers) UploadGame(c *gin.Context) {
	var req UploadGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}

	pgn, err := game.ParsePGN(req.PGN)
	if err != nil {
		c.JSON(400, gin.H{"error": "failed to parse PGN"})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	g := &models.Game{
		UserID:      userID,
		PGN:         req.PGN,
		WhitePlayer: pgn.WhitePlayer,
		BlackPlayer: pgn.BlackPlayer,
		Result:      pgn.Result,
		Opening:     pgn.Opening,
		Event:       pgn.Event,
		Date:        pgn.Date,
	}

	if err := h.gameRepo.Create(context.Background(), g); err != nil {
		c.JSON(500, gin.H{"error": "failed to save game"})
		return
	}

	c.JSON(201, gin.H{
		"message": "game uploaded successfully",
		"game_id": g.ID,
	})
}

func (h *GameHandlers) ListGames(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	games, err := h.gameRepo.ListByUserID(context.Background(), userID, 50, 0)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to list games"})
		return
	}

	c.JSON(200, gin.H{"games": games})
}

func (h *GameHandlers) GetGame(c *gin.Context) {
	id := c.Param("id")
	
	game, err := h.gameRepo.GetByID(context.Background(), id)
	if err != nil {
		c.JSON(404, gin.H{"error": "game not found"})
		return
	}

	c.JSON(200, gin.H{"game": game})
}

func (h *GameHandlers) DeleteGame(c *gin.Context) {
	id := c.Param("id")
	
	if err := h.gameRepo.Delete(context.Background(), id); err != nil {
		c.JSON(500, gin.H{"error": "failed to delete game"})
		return
	}

	c.JSON(200, gin.H{"message": "game deleted"})
}

type CreateAnalysisRequest struct {
	GameID string `json:"game_id" binding:"required"`
	Depth  int    `json:"depth"`
}

func (h *GameHandlers) CreateAnalysis(c *gin.Context) {
	var req CreateAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	game, err := h.gameRepo.GetByID(context.Background(), req.GameID)
	if err != nil {
		c.JSON(404, gin.H{"error": "game not found"})
		return
	}

	depth := req.Depth
	if depth == 0 {
		depth = 20
	}

	session := &models.AnalysisSession{
		GameID:       req.GameID,
		UserID:       userID,
		Depth:        depth,
		Status:       "pending",
		EngineConfig: `{"threads": 1, "hash": 256}`,
	}

	if err := h.sessionRepo.Create(context.Background(), session); err != nil {
		c.JSON(500, gin.H{"error": "failed to create analysis session"})
		return
	}

	// Start analysis asynchronously
	if h.analysisService != nil {
		go func() {
			ctx := context.Background()
			if err := h.analysisService.AnalyzeGame(ctx, session, game.PGN); err != nil {
				h.sessionRepo.UpdateStatus(ctx, session.ID, "failed")
			}
		}()
	}

	c.JSON(201, gin.H{
		"message":    "analysis session created",
		"session_id": session.ID,
		"status":     "pending",
	})
}

func (h *GameHandlers) GetAnalysis(c *gin.Context) {
	id := c.Param("id")
	
	session, err := h.sessionRepo.GetByID(context.Background(), id)
	if err != nil {
		c.JSON(404, gin.H{"error": "analysis session not found"})
		return
	}

	c.JSON(200, gin.H{"session": session})
}
