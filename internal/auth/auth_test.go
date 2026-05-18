package auth

import (
	"testing"

	"github.com/chessgoddess/chesslens/internal/config"
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
