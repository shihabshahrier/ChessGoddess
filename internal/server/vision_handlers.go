package server

import (
	"io"
	"net/http"

	"github.com/chessgoddess/chesslens/internal/vision"
	"github.com/gin-gonic/gin"
)

type VisionHandlers struct {
	visionClient *vision.ImageToFENClient
}

func NewVisionHandlers(visionClient *vision.ImageToFENClient) *VisionHandlers {
	return &VisionHandlers{
		visionClient: visionClient,
	}
}

func (h *VisionHandlers) ImageToFEN(c *gin.Context) {
	file, _, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(400, gin.H{"error": "image file required"})
		return
	}
	defer file.Close()

	imageData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to read image"})
		return
	}

	fen, err := h.visionClient.ImageToFEN(c.Request.Context(), imageData)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to extract FEN from image"})
		return
	}

	c.JSON(200, gin.H{"fen": fen})
}

func (h *VisionHandlers) ImageToFENURL(c *gin.Context) {
	var req struct {
		ImageURL string `json:"image_url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "image_url required"})
		return
	}

	resp, err := http.Get(req.ImageURL)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch image"})
		return
	}
	defer resp.Body.Close()

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to read image"})
		return
	}

	fen, err := h.visionClient.ImageToFEN(c.Request.Context(), imageData)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to extract FEN from image"})
		return
	}

	c.JSON(200, gin.H{"fen": fen})
}
