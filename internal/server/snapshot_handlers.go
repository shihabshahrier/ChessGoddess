package server

import (
	"context"

	"github.com/chessgoddess/chesslens/internal/repository"
	"github.com/gin-gonic/gin"
)

type SnapshotHandlers struct {
	snapshotRepo *repository.SnapshotRepository
}

func NewSnapshotHandlers(snapshotRepo *repository.SnapshotRepository) *SnapshotHandlers {
	return &SnapshotHandlers{
		snapshotRepo: snapshotRepo,
	}
}

func (h *SnapshotHandlers) GetByShareToken(c *gin.Context) {
	token := c.Param("token")
	
	snapshot, err := h.snapshotRepo.GetByShareToken(context.Background(), token)
	if err != nil {
		c.JSON(404, gin.H{"error": "snapshot not found"})
		return
	}

	c.JSON(200, gin.H{"snapshot": snapshot})
}

func (h *SnapshotHandlers) ListByUser(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	snapshots, err := h.snapshotRepo.ListByUserID(context.Background(), userID, 50, 0)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to list snapshots"})
		return
	}

	c.JSON(200, gin.H{"snapshots": snapshots})
}
