package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chessgoddess/chessgoddess/internal/config"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// --- AI handlers ---

func TestAIHandlers_NilService_ExplainMove(t *testing.T) {
	h := NewAIHandlers(nil, nil)

	r := gin.New()
	r.POST("/ai/explain", h.ExplainMove)

	body := bytes.NewBufferString(`{"move_id":"m1","session_id":"s1"}`)
	req := httptest.NewRequest(http.MethodPost, "/ai/explain", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
}

func TestAIHandlers_BadBody_ExplainMove(t *testing.T) {
	h := NewAIHandlers(nil, nil)

	r := gin.New()
	r.POST("/ai/explain", h.ExplainMove)

	req := httptest.NewRequest(http.MethodPost, "/ai/explain", bytes.NewBufferString(`{bad json}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAIHandlers_NilService_ExplainBlunder(t *testing.T) {
	h := NewAIHandlers(nil, nil)

	r := gin.New()
	r.POST("/ai/explain-blunder", h.ExplainBlunder)

	body := bytes.NewBufferString(`{"move_id":"m1","session_id":"s1"}`)
	req := httptest.NewRequest(http.MethodPost, "/ai/explain-blunder", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
}

func TestAIHandlers_BadBody_ExplainBlunder(t *testing.T) {
	h := NewAIHandlers(nil, nil)

	r := gin.New()
	r.POST("/ai/explain-blunder", h.ExplainBlunder)

	req := httptest.NewRequest(http.MethodPost, "/ai/explain-blunder", bytes.NewBufferString(`not json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAIHandlers_NilService_GetExplanation(t *testing.T) {
	h := NewAIHandlers(nil, nil)

	r := gin.New()
	r.GET("/ai/explanation/:move_id", h.GetExplanation)

	req := httptest.NewRequest(http.MethodGet, "/ai/explanation/m1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
}

// --- Game handlers ---

func TestGameHandlers_BadBody_UploadGame(t *testing.T) {
	h := NewGameHandlers(nil, nil, nil, nil)

	r := gin.New()
	r.POST("/games/upload", h.UploadGame)

	req := httptest.NewRequest(http.MethodPost, "/games/upload", bytes.NewBufferString(`not json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGameHandlers_NoUserID_UploadGame(t *testing.T) {
	h := NewGameHandlers(nil, nil, nil, nil)

	r := gin.New()
	r.POST("/games/upload", h.UploadGame)

	body := bytes.NewBufferString(`{"pgn":"1. e4 e5"}`)
	req := httptest.NewRequest(http.MethodPost, "/games/upload", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Fails at PGN parse (game.ParsePGN) or user_id check — either 400 or 401.
	if w.Code != http.StatusBadRequest && w.Code != http.StatusUnauthorized {
		t.Errorf("expected 400 or 401, got %d", w.Code)
	}
}

func TestGameHandlers_NoUserID_ListGames(t *testing.T) {
	h := NewGameHandlers(nil, nil, nil, nil)

	r := gin.New()
	r.GET("/games", h.ListGames)

	req := httptest.NewRequest(http.MethodGet, "/games", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestGameHandlers_BadBody_CreateAnalysis(t *testing.T) {
	h := NewGameHandlers(nil, nil, nil, nil)

	r := gin.New()
	r.POST("/analysis", h.CreateAnalysis)

	req := httptest.NewRequest(http.MethodPost, "/analysis", bytes.NewBufferString(`not json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGameHandlers_NoUserID_CreateAnalysis(t *testing.T) {
	h := NewGameHandlers(nil, nil, nil, nil)

	r := gin.New()
	r.POST("/analysis", h.CreateAnalysis)

	body := bytes.NewBufferString(`{"game_id":"g1"}`)
	req := httptest.NewRequest(http.MethodPost, "/analysis", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// --- Snapshot handlers ---

func TestSnapshotHandlers_NoUserID_ListByUser(t *testing.T) {
	h := NewSnapshotHandlers(nil)

	r := gin.New()
	r.GET("/snapshots", h.ListByUser)

	req := httptest.NewRequest(http.MethodGet, "/snapshots", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// --- Auth handlers ---

func TestAuthHandlers_GoogleCallback_NoState(t *testing.T) {
	h := NewAuthHandlers(nil, nil, &config.Config{FrontendURL: "http://localhost:3000"})

	r := gin.New()
	r.GET("/auth/google/callback", h.GoogleCallback)

	req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?state=mismatch&code=abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// oauth_state cookie missing → bad request
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAuthHandlers_GoogleCallback_MissingCode(t *testing.T) {
	h := NewAuthHandlers(nil, nil, &config.Config{FrontendURL: "http://localhost:3000"})

	r := gin.New()
	r.GET("/auth/google/callback", h.GoogleCallback)

	// Set matching state cookie but no code param.
	req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?state=abc123", nil)
	req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "abc123"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
