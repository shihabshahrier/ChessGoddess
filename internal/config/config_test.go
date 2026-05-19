package config

import (
	"testing"
)

// --- parseOrigins ---

func TestParseOrigins_Single(t *testing.T) {
	got := parseOrigins("https://example.com")
	if len(got) != 1 || got[0] != "https://example.com" {
		t.Errorf("unexpected: %v", got)
	}
}

func TestParseOrigins_Multiple(t *testing.T) {
	got := parseOrigins("https://a.com, https://b.com , https://c.com")
	if len(got) != 3 {
		t.Fatalf("expected 3, got %d: %v", len(got), got)
	}
	if got[0] != "https://a.com" || got[1] != "https://b.com" || got[2] != "https://c.com" {
		t.Errorf("unexpected values: %v", got)
	}
}

func TestParseOrigins_Empty(t *testing.T) {
	got := parseOrigins("")
	if len(got) != 0 {
		t.Errorf("expected 0, got %d", len(got))
	}
}

func TestParseOrigins_Whitespace(t *testing.T) {
	got := parseOrigins("  , ,  ")
	if len(got) != 0 {
		t.Errorf("expected 0, got %d: %v", len(got), got)
	}
}

// --- Validate ---

func TestValidate_MissingClientID(t *testing.T) {
	cfg := &Config{
		GoogleClientID: "",
		GoogleSecret:   "secret",
		JWTSecret:      "jwt",
		Environment:    "development",
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for missing GOOGLE_CLIENT_ID")
	}
}

func TestValidate_MissingClientSecret(t *testing.T) {
	cfg := &Config{
		GoogleClientID: "id",
		GoogleSecret:   "",
		JWTSecret:      "jwt",
		Environment:    "development",
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for missing GOOGLE_CLIENT_SECRET")
	}
}

func TestValidate_ProductionWeakJWT(t *testing.T) {
	cfg := &Config{
		GoogleClientID: "id",
		GoogleSecret:   "secret",
		JWTSecret:      "dev-secret-change-in-production",
		Environment:    "production",
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for weak JWT_SECRET in production")
	}
}

func TestValidate_ProductionStrongJWT(t *testing.T) {
	cfg := &Config{
		GoogleClientID: "id",
		GoogleSecret:   "secret",
		JWTSecret:      "a-very-secure-secret-for-production",
		Environment:    "production",
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_DevelopmentWeakJWTAllowed(t *testing.T) {
	cfg := &Config{
		GoogleClientID: "id",
		GoogleSecret:   "secret",
		JWTSecret:      "dev-secret-change-in-production",
		Environment:    "development",
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error in dev with weak secret: %v", err)
	}
}

// --- GoogleRedirectURL ---

func TestGoogleRedirectURL_Default(t *testing.T) {
	cfg := &Config{}
	got := cfg.GoogleRedirectURL()
	want := "http://localhost:8080/api/v1/auth/google/callback"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestGoogleRedirectURL_EnvOverride(t *testing.T) {
	t.Setenv("GOOGLE_REDIRECT_URL", "https://prod.example.com/callback")
	cfg := &Config{}
	got := cfg.GoogleRedirectURL()
	if got != "https://prod.example.com/callback" {
		t.Errorf("expected env override, got %q", got)
	}
}

// --- getEnv ---

func TestGetEnv_UsesEnvVar(t *testing.T) {
	t.Setenv("TEST_GET_ENV_KEY", "fromenv")
	got := getEnv("TEST_GET_ENV_KEY", "fallback")
	if got != "fromenv" {
		t.Errorf("expected fromenv, got %q", got)
	}
}

func TestGetEnv_UsesFallback(t *testing.T) {
	got := getEnv("TEST_KEY_THAT_DOES_NOT_EXIST_XYZ", "fallback")
	if got != "fallback" {
		t.Errorf("expected fallback, got %q", got)
	}
}
