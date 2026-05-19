// Vision handlers — chess board image to FEN conversion endpoints.
package api

import (
	"io"
	"net/http"

	"github.com/chessgoddess/chesslens/internal/service"
	"github.com/gin-gonic/gin"
)

type VisionHandlers struct {
	visionClient *service.VisionClient
}

func NewVisionHandlers(visionClient *service.VisionClient) *VisionHandlers {
	return &VisionHandlers{visionClient: visionClient}
}

func (h *VisionHandlers) ImageToFEN(c *gin.Context) {
	if h.visionClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "vision service not configured"})
		return
	}

	file, _, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image file required"})
		return
	}
	defer file.Close()

	imageData, err := io.ReadAll(io.LimitReader(file, 10<<20)) // 10 MB limit
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read image"})
		return
	}

	fen, err := h.visionClient.ImageToFEN(c.Request.Context(), imageData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to extract FEN from image"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"fen": fen})
}

func (h *VisionHandlers) ImageToFENURL(c *gin.Context) {
	if h.visionClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "vision service not configured"})
		return
	}

	var req struct {
		ImageURL string `json:"image_url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image_url required"})
		return
	}

	httpReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, req.ImageURL, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid image URL"})
		return
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch image"})
		return
	}
	defer resp.Body.Close()

	imageData, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read image"})
		return
	}

	fen, err := h.visionClient.ImageToFEN(c.Request.Context(), imageData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to extract FEN from image"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"fen": fen})
}
