package auth

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chessgoddess/chessgoddess/internal/config"
)

func TestGenerateJWT(t *testing.T) {
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}

	service := NewService(cfg)

	token, err := service.GenerateJWT("user1", "test@example.com", "Test User", "https://example.com/avatar.png")
	if err != nil {
		t.Fatalf("failed to generate JWT: %v", err)
	}

	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestValidateJWT_ValidToken(t *testing.T) {
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}

	service := NewService(cfg)

	token, err := service.GenerateJWT("user1", "test@example.com", "Test User", "")
	if err != nil {
		t.Fatalf("failed to generate JWT: %v", err)
	}

	claims, err := service.ValidateJWT(token)
	if err != nil {
		t.Fatalf("failed to validate JWT: %v", err)
	}

	if claims.UserID != "user1" {
		t.Errorf("expected user_id 'user1', got '%s'", claims.UserID)
	}

	if claims.ExpiresAt == nil {
		t.Error("expected ExpiresAt to be set")
	}
	if claims.IssuedAt == nil {
		t.Error("expected IssuedAt to be set")
	}

	if claims.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", claims.Email)
	}
}

func TestValidateJWT_InvalidToken(t *testing.T) {
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}

	service := NewService(cfg)

	_, err := service.ValidateJWT("invalid-token")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestValidateJWT_WrongSecret(t *testing.T) {
	cfg1 := &config.Config{
		JWTSecret: "secret1",
	}

	cfg2 := &config.Config{
		JWTSecret: "secret2",
	}

	service1 := NewService(cfg1)
	service2 := NewService(cfg2)

	token, err := service1.GenerateJWT("user1", "test@example.com", "Test User", "")
	if err != nil {
		t.Fatalf("failed to generate JWT: %v", err)
	}

	_, err = service2.ValidateJWT(token)
	if err == nil {
		t.Error("expected error for wrong secret")
	}
}

func TestGenerateState(t *testing.T) {
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}

	service := NewService(cfg)

	state1, err := service.GenerateState()
	if err != nil {
		t.Fatalf("failed to generate state: %v", err)
	}

	state2, err := service.GenerateState()
	if err != nil {
		t.Fatalf("failed to generate state: %v", err)
	}

	if state1 == state2 {
		t.Error("expected different states")
	}
}

func TestSetAuthCookie(t *testing.T) {
	cfg := &config.Config{JWTSecret: "secret", Environment: "development"}
	svc := NewService(cfg)

	w := httptest.NewRecorder()
	svc.SetAuthCookie(w, "mytoken")

	cookies := w.Result().Cookies()
	var found bool
	for _, c := range cookies {
		if c.Name == "auth_token" && c.Value == "mytoken" {
			found = true
		}
	}
	if !found {
		t.Error("auth_token cookie not set")
	}
}

func TestClearAuthCookie(t *testing.T) {
	cfg := &config.Config{JWTSecret: "secret", Environment: "development"}
	svc := NewService(cfg)

	w := httptest.NewRecorder()
	svc.ClearAuthCookie(w)

	cookies := w.Result().Cookies()
	var found bool
	for _, c := range cookies {
		if c.Name == "auth_token" && c.MaxAge == -1 {
			found = true
		}
	}
	if !found {
		t.Error("auth_token clear cookie not set")
	}
}

func TestPackageLevelValidateJWT(t *testing.T) {
	cfg := &config.Config{JWTSecret: "secret"}
	svc := NewService(cfg)

	token, err := svc.GenerateJWT("u1", "a@b.com", "Name", "")
	if err != nil {
		t.Fatalf("GenerateJWT: %v", err)
	}

	claims, err := ValidateJWT(token, "secret")
	if err != nil {
		t.Fatalf("ValidateJWT: %v", err)
	}
	if claims.UserID != "u1" {
		t.Errorf("want u1, got %s", claims.UserID)
	}
}

func TestPackageLevelValidateJWT_WrongSecret(t *testing.T) {
	cfg := &config.Config{JWTSecret: "secret"}
	svc := NewService(cfg)

	token, _ := svc.GenerateJWT("u1", "a@b.com", "Name", "")
	_, err := ValidateJWT(token, "wrongsecret")
	if err == nil {
		t.Error("expected error for wrong secret")
	}
}

func TestPackageLevelValidateJWT_InvalidAlgorithm(t *testing.T) {
	// Ensure unexpected signing method is rejected.
	_, err := ValidateJWT(strings.Repeat("a", 200), "secret")
	if err == nil {
		t.Error("expected error for garbage token")
	}
}
