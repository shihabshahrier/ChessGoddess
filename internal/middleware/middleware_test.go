package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newRouter(handlers ...gin.HandlerFunc) *gin.Engine {
	r := gin.New()
	r.Use(handlers...)
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return r
}

// --- CORS ---

func TestCORS_AllowedOrigin(t *testing.T) {
	r := newRouter(CORS([]string{"https://example.com"}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Errorf("expected ACAO header to be set, got %q", w.Header().Get("Access-Control-Allow-Origin"))
	}
	if w.Header().Get("Vary") != "Origin" {
		t.Error("expected Vary: Origin header")
	}
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	r := newRouter(CORS([]string{"https://example.com"}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://evil.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("expected no ACAO header for disallowed origin")
	}
}

func TestCORS_Preflight(t *testing.T) {
	r := newRouter(CORS([]string{"https://example.com"}))
	r.OPTIONS("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("preflight: expected 204, got %d", w.Code)
	}
}

func TestCORS_NoOriginHeader(t *testing.T) {
	r := newRouter(CORS([]string{"https://example.com"}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for no-origin request, got %d", w.Code)
	}
}

// --- RateLimiter ---

func TestRateLimiter_AllowsUnderLimit(t *testing.T) {
	r := newRouter(RateLimiter(100, 100))

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "127.0.0.1:1234"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i, w.Code)
		}
	}
}

func TestRateLimiter_BlocksOverLimit(t *testing.T) {
	// 1 token per second, burst 1 — second request should be blocked.
	r := newRouter(RateLimiter(rate.Limit(1), 1))

	send := func() int {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "10.0.0.1:9999"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w.Code
	}

	first := send()
	if first != http.StatusOK {
		t.Errorf("first request should pass, got %d", first)
	}

	second := send()
	if second != http.StatusTooManyRequests {
		t.Errorf("second burst request should be rate-limited, got %d", second)
	}
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	// Burst 1 per IP — IP A exhausted should not affect IP B.
	r := newRouter(RateLimiter(rate.Limit(1), 1))

	sendFrom := func(ip string) int {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = ip + ":1234"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w.Code
	}

	sendFrom("192.168.1.1") // consume IP A's burst
	sendFrom("192.168.1.1") // should be blocked

	if code := sendFrom("192.168.1.2"); code != http.StatusOK {
		t.Errorf("different IP should not be rate-limited, got %d", code)
	}
}

// --- Logger ---

func TestLogger_DoesNotPanic(t *testing.T) {
	r := newRouter(Logger())

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Logger panicked: %v", r)
		}
	}()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// --- Recovery ---

func TestRecovery_CatchesPanic(t *testing.T) {
	r := gin.New()
	r.Use(Recovery())
	r.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 after panic, got %d", w.Code)
	}
}

// --- Auth ---

func TestAuth_MissingHeader(t *testing.T) {
	r := newRouter(Auth("secret"))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuth_InvalidFormat(t *testing.T) {
	r := newRouter(Auth("secret"))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "InvalidToken")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuth_InvalidToken(t *testing.T) {
	r := newRouter(Auth("secret"))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer notavalidjwt")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// ensure time import is used (rate limiter cleanup ticker)
var _ = time.Minute
