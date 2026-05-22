// Game handlers — upload, analysis, and game management endpoints.
package api

import (
	"context"
	"net/http"

	"github.com/chessgoddess/chessgoddess/internal/game"
	"github.com/chessgoddess/chessgoddess/internal/model"
	"github.com/chessgoddess/chessgoddess/internal/repository"
	"github.com/chessgoddess/chessgoddess/internal/service"
	"github.com/chessgoddess/chessgoddess/internal/worker"
	"github.com/gin-gonic/gin"
)

type GameHandlers struct {
	gameRepo        *repository.GameRepository
	sessionRepo     *repository.AnalysisSessionRepository
	analysisService *service.AnalysisService
	queue           worker.JobQueue
}

func NewGameHandlers(
	gameRepo *repository.GameRepository,
	sessionRepo *repository.AnalysisSessionRepository,
	analysisSvc *service.AnalysisService,
	q worker.JobQueue,
) *GameHandlers {
	return &GameHandlers{
		gameRepo:        gameRepo,
		sessionRepo:     sessionRepo,
		analysisService: analysisSvc,
		queue:           q,
	}
}

type uploadGameRequest struct {
	PGN string `json:"pgn" binding:"required"`
}

func (h *GameHandlers) UploadGame(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20) // 1 MB limit

	var req uploadGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	pgn, err := game.ParsePGN(req.PGN)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse PGN"})
		return
	}

	g := &model.Game{
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save game"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "game uploaded successfully", "game_id": g.ID})
}

func (h *GameHandlers) ListGames(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	games, err := h.gameRepo.ListByUserID(context.Background(), userID, 50, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list games"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"games": games})
}

func (h *GameHandlers) GetGame(c *gin.Context) {
	g, err := h.gameRepo.GetByID(context.Background(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"game": g})
}

func (h *GameHandlers) DeleteGame(c *gin.Context) {
	if err := h.gameRepo.Delete(context.Background(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete game"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "game deleted"})
}

type createAnalysisRequest struct {
	GameID string `json:"game_id" binding:"required"`
	Depth  int    `json:"depth"`
}

func (h *GameHandlers) CreateAnalysis(c *gin.Context) {
	var req createAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	g, err := h.gameRepo.GetByID(context.Background(), req.GameID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	depth := req.Depth
	if depth == 0 {
		depth = 20
	}

	session := &model.AnalysisSession{
		GameID:       g.ID,
		UserID:       userID,
		Depth:        depth,
		Status:       "pending",
		EngineConfig: `{"threads": 1, "hash": 256}`,
	}

	if err := h.sessionRepo.Create(context.Background(), session); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create analysis session"})
		return
	}

	if h.queue != nil {
		if err := h.queue.EnqueueAnalysis(session.ID, g.ID, depth); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enqueue analysis job"})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "analysis session created",
		"session_id": session.ID,
		"status":     "pending",
	})
}

func (h *GameHandlers) GetAnalysis(c *gin.Context) {
	session, err := h.sessionRepo.GetByID(context.Background(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "analysis session not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"session": session})
}

func (h *GameHandlers) GetAnalysisMoves(c *gin.Context) {
	if h.analysisService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "analysis service unavailable"})
		return
	}
	moves, err := h.analysisService.GetMovesBySessionID(context.Background(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load moves"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"moves": moves})
}

func (h *GameHandlers) CreateSnapshot(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if h.queue != nil {
		if err := h.queue.EnqueueSnapshot(sessionID, userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enqueue snapshot job"})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{"message": "snapshot creation queued", "session_id": sessionID})
}
