package server

import (
	"context"

	"github.com/chessgoddess/chesslens/internal/ai"
	"github.com/chessgoddess/chesslens/internal/repository"
	"github.com/gin-gonic/gin"
)

type AIHandlers struct {
	aiService *ai.Service
	moveRepo  *repository.MoveRepository
}

func NewAIHandlers(aiService *ai.Service, moveRepo *repository.MoveRepository) *AIHandlers {
	return &AIHandlers{
		aiService: aiService,
		moveRepo:  moveRepo,
	}
}

type ExplainMoveRequest struct {
	MoveID string `json:"move_id" binding:"required"`
}

func (h *AIHandlers) ExplainMove(c *gin.Context) {
	var req ExplainMoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}

	move, err := h.moveRepo.GetByID(context.Background(), req.MoveID)
	if err != nil {
		c.JSON(404, gin.H{"error": "move not found"})
		return
	}

	sessionID := c.GetString("session_id")
	if sessionID == "" {
		c.JSON(400, gin.H{"error": "session_id required"})
		return
	}

	explanation, err := h.aiService.ExplainMove(
		context.Background(),
		sessionID,
		move.ID,
		move.FEN,
		move.SAN,
		move.Classification,
		move.Evaluation,
	)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to generate explanation"})
		return
	}

	c.JSON(200, gin.H{"explanation": explanation})
}

type ExplainBlunderRequest struct {
	MoveID string `json:"move_id" binding:"required"`
}

func (h *AIHandlers) ExplainBlunder(c *gin.Context) {
	var req ExplainBlunderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}

	move, err := h.moveRepo.GetByID(context.Background(), req.MoveID)
	if err != nil {
		c.JSON(404, gin.H{"error": "move not found"})
		return
	}

	sessionID := c.GetString("session_id")
	if sessionID == "" {
		c.JSON(400, gin.H{"error": "session_id required"})
		return
	}

	explanation, err := h.aiService.ExplainBlunder(
		context.Background(),
		sessionID,
		move.ID,
		move.FEN,
		move.SAN,
		move.BestMove,
		0,
		move.Evaluation,
	)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to generate blunder explanation"})
		return
	}

	c.JSON(200, gin.H{"explanation": explanation})
}

func (h *AIHandlers) GetExplanation(c *gin.Context) {
	moveID := c.Param("move_id")

	explanation, err := h.aiService.GetRepo().GetByMoveID(context.Background(), moveID)
	if err != nil {
		c.JSON(404, gin.H{"error": "explanation not found"})
		return
	}

	c.JSON(200, gin.H{"explanation": explanation})
}
