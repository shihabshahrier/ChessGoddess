// AI handlers — move explanation and blunder analysis endpoints.
package api

import (
	"context"
	"net/http"

	"github.com/chessgoddess/chessgoddess/internal/repository"
	"github.com/chessgoddess/chessgoddess/internal/service"
	"github.com/gin-gonic/gin"
)

type AIHandlers struct {
	aiService *service.AIService
	moveRepo  *repository.MoveRepository
}

func NewAIHandlers(aiService *service.AIService, moveRepo *repository.MoveRepository) *AIHandlers {
	return &AIHandlers{aiService: aiService, moveRepo: moveRepo}
}

type explainMoveRequest struct {
	MoveID    string `json:"move_id" binding:"required"`
	SessionID string `json:"session_id" binding:"required"`
}

func (h *AIHandlers) ExplainMove(c *gin.Context) {
	var req explainMoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if h.aiService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI service not configured"})
		return
	}

	move, err := h.moveRepo.GetByID(context.Background(), req.MoveID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "move not found"})
		return
	}

	explanation, err := h.aiService.ExplainMove(
		context.Background(),
		req.SessionID,
		move.ID,
		move.FEN,
		move.SAN,
		move.Classification,
		move.Evaluation,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate explanation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"explanation": explanation})
}

type explainBlunderRequest struct {
	MoveID    string `json:"move_id" binding:"required"`
	SessionID string `json:"session_id" binding:"required"`
}

func (h *AIHandlers) ExplainBlunder(c *gin.Context) {
	var req explainBlunderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if h.aiService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI service not configured"})
		return
	}

	move, err := h.moveRepo.GetByID(context.Background(), req.MoveID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "move not found"})
		return
	}

	explanation, err := h.aiService.ExplainBlunder(
		context.Background(),
		req.SessionID,
		move.ID,
		move.FEN,
		move.SAN,
		move.BestMove,
		0,
		move.Evaluation,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate blunder explanation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"explanation": explanation})
}

func (h *AIHandlers) GetExplanation(c *gin.Context) {
	if h.aiService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI service not configured"})
		return
	}

	explanation, err := h.aiService.GetRepo().GetByMoveID(context.Background(), c.Param("move_id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "explanation not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"explanation": explanation})
}
