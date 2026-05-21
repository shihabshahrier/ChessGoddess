// Auth handlers — Google OAuth2 flow.
package api

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/chessgoddess/chesslens/internal/auth"
	"github.com/chessgoddess/chesslens/internal/config"
	"github.com/chessgoddess/chesslens/internal/model"
	"github.com/chessgoddess/chesslens/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type AuthHandlers struct {
	service  *auth.Service
	userRepo *repository.UserRepository
	config   *config.Config
}

func NewAuthHandlers(service *auth.Service, userRepo *repository.UserRepository, cfg *config.Config) *AuthHandlers {
	return &AuthHandlers{service: service, userRepo: userRepo, config: cfg}
}

func (h *AuthHandlers) GetGoogleAuthURL(c *gin.Context) {
	state, err := h.service.GenerateState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate state"})
		return
	}

	secure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		MaxAge:   3600,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
	c.JSON(http.StatusOK, gin.H{"url": h.service.GetAuthURL(state)})
}

func (h *AuthHandlers) GoogleCallback(c *gin.Context) {
	// Validate state cookie (CSRF protection).
	stateCookie, err := c.Cookie("oauth_state")
	if err != nil || stateCookie != c.Query("state") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid oauth state"})
		return
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
		return
	}

	token, err := h.service.ExchangeCode(c.Request.Context(), code)
	if err != nil {
		slog.Error("oauth code exchange failed", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to exchange code"})
		return
	}

	// Fetch Google user info.
	client := h.service.GetHTTPClient(c.Request.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		slog.Error("failed to fetch google userinfo", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user info"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("google userinfo returned non-200", "status", resp.StatusCode)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user info"})
		return
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read user info"})
		return
	}

	var googleUser struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		Name      string `json:"name"`
		Picture   string `json:"picture"`
	}
	if err := json.Unmarshal(body, &googleUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse user info"})
		return
	}

	if googleUser.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "google account has no email"})
		return
	}

	// Upsert user.
	user, err := h.userRepo.GetByGoogleID(c.Request.Context(), googleUser.ID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			slog.Error("db lookup failed", "err", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
		user = &model.User{
			GoogleID:  googleUser.ID,
			Email:     googleUser.Email,
			Name:      googleUser.Name,
			AvatarURL: googleUser.Picture,
		}
		if err := h.userRepo.Create(c.Request.Context(), user); err != nil {
			slog.Error("failed to create user", "err", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}
	}

	jwtToken, err := h.service.GenerateJWT(user.ID, user.Email, user.Name, user.AvatarURL)
	if err != nil {
		slog.Error("jwt generation failed", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	h.service.SetAuthCookie(c.Writer, jwtToken)
	c.Redirect(http.StatusFound, h.config.FrontendURL+"/dashboard")
}
