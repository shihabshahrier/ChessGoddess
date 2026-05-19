// Snapshot handlers — public share links and user snapshot listing.
package api

import (
	"context"
	"net/http"

	"github.com/chessgoddess/chesslens/internal/repository"
	"github.com/gin-gonic/gin"
)

type SnapshotHandlers struct {
	snapshotRepo *repository.SnapshotRepository
}

func NewSnapshotHandlers(snapshotRepo *repository.SnapshotRepository) *SnapshotHandlers {
	return &SnapshotHandlers{snapshotRepo: snapshotRepo}
}

func (h *SnapshotHandlers) GetByShareToken(c *gin.Context) {
	snapshot, err := h.snapshotRepo.GetByShareToken(context.Background(), c.Param("token"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "snapshot not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"snapshot": snapshot})
}

func (h *SnapshotHandlers) ListByUser(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	snapshots, err := h.snapshotRepo.ListByUserID(context.Background(), userID, 50, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list snapshots"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"snapshots": snapshots})
}
