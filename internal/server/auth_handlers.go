package server

import (
	"github.com/chessgoddess/chesslens/internal/auth"
	"github.com/gin-gonic/gin"
)

type AuthHandlers struct {
	service *auth.Service
}

func NewAuthHandlers(service *auth.Service) *AuthHandlers {
	return &AuthHandlers{
		service: service,
	}
}

func (h *AuthHandlers) GetGoogleAuthURL(c *gin.Context) {
	state, err := h.service.GenerateState()
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to generate state"})
		return
	}

	c.SetCookie("oauth_state", state, 3600, "/", "", false, true)
	url := h.service.GetAuthURL(state)
	
	c.JSON(200, gin.H{"url": url})
}

func (h *AuthHandlers) GoogleCallback(c *gin.Context) {
	// TODO: Implement full OAuth callback with user repository
	// For now, return placeholder
	c.JSON(200, gin.H{
		"message": "OAuth callback - implement with user repository",
		"code":    c.Query("code"),
		"state":   c.Query("state"),
	})
}
