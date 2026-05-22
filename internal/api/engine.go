// Engine handlers — on-demand position evaluation for the analysis board.
package api

import (
	"net/http"

	"github.com/chessgoddess/chessgoddess/internal/engine"
	"github.com/gin-gonic/gin"
	"github.com/notnil/chess"
)

type EngineHandlers struct {
	pool *engine.Pool
}

func NewEngineHandlers(pool *engine.Pool) *EngineHandlers {
	return &EngineHandlers{pool: pool}
}

type evaluateRequest struct {
	FEN     string `json:"fen" binding:"required"`
	Depth   int    `json:"depth"`
	MultiPV int    `json:"multipv"`
}

// Evaluate runs a bounded engine search on a single position. The board's
// continuous eval runs client-side (WASM); this is for deeper on-demand dives.
func (h *EngineHandlers) Evaluate(c *gin.Context) {
	if h.pool == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "engine not available"})
		return
	}

	var req evaluateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fen is required"})
		return
	}
	if _, err := chess.FEN(req.FEN); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid FEN"})
		return
	}

	// Clamp depth and lines to protect server CPU.
	depth := req.Depth
	if depth < 1 || depth > 22 {
		depth = 14
	}
	multipv := req.MultiPV
	if multipv < 1 || multipv > 5 {
		multipv = 3
	}

	eval, err := h.pool.Evaluate(c.Request.Context(), req.FEN, depth, multipv)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "evaluation failed"})
		return
	}

	c.JSON(http.StatusOK, eval)
}
