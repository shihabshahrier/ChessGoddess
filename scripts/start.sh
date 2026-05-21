#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PIDFILE="$ROOT/.dev.pids"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

ok()   { echo -e "  ${GREEN}✓${NC} $1"; }
fail() { echo -e "  ${RED}✗${NC} $1"; }
warn() { echo -e "  ${YELLOW}!${NC} $1"; }

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " ChessGoddess — Local Dev Startup"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# ── Load .env ──
if [ ! -f "$ROOT/.env" ]; then
    fail ".env file not found. Copy .env.example and fill in values."
    exit 1
fi
set -a; source "$ROOT/.env"; set +a
ok ".env loaded"

# ── Check if already running ──
if [ -f "$PIDFILE" ]; then
    warn "Dev servers may already be running. Run ./scripts/stop.sh first."
    exit 1
fi

# ── Preflight checks ──
echo ""
echo "Preflight checks:"
CHECKS_PASSED=true

# Go
if command -v go &>/dev/null; then
    ok "Go $(go version | awk '{print $3}' | sed 's/go//')"
else
    fail "Go not installed"; CHECKS_PASSED=false
fi

# Node
if command -v node &>/dev/null; then
    ok "Node $(node -v)"
else
    fail "Node not installed"; CHECKS_PASSED=false
fi

# npm
if command -v npm &>/dev/null; then
    ok "npm $(npm -v)"
else
    fail "npm not installed"; CHECKS_PASSED=false
fi

# Stockfish (optional)
if command -v "${STOCKFISH_PATH:-stockfish}" &>/dev/null; then
    ok "Stockfish found at ${STOCKFISH_PATH:-stockfish}"
else
    warn "Stockfish not found — analysis will be unavailable"
fi

# Database connectivity
if go run -C "$ROOT" ./cmd/server 2>&1 | head -1 | grep -q "failed"; then
    : # skip, we'll test directly
fi
echo ""
echo "Service connectivity:"

# PostgreSQL
if psql "$DATABASE_URL" -c "SELECT 1" &>/dev/null 2>&1; then
    ok "PostgreSQL reachable"
elif pg_isready -d "$DATABASE_URL" &>/dev/null 2>&1; then
    ok "PostgreSQL reachable"
else
    # Try Go-based ping
    if go run "$ROOT/scripts/healthcheck.go" db 2>/dev/null; then
        ok "PostgreSQL reachable"
    else
        warn "PostgreSQL — cannot verify (will check on startup)"
    fi
fi

# Redis
if command -v redis-cli &>/dev/null; then
    if redis-cli -u "$REDIS_URL" ping 2>/dev/null | grep -q "PONG"; then
        ok "Redis reachable"
    else
        warn "Redis — cannot connect (will check on startup)"
    fi
else
    warn "Redis — redis-cli not installed, skipping check"
fi

# OpenRouter API
if [ -n "${OPENROUTER_API_KEY:-}" ]; then
    STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
        -H "Authorization: Bearer $OPENROUTER_API_KEY" \
        "https://openrouter.ai/api/v1/models" 2>/dev/null || echo "000")
    if [ "$STATUS" = "200" ]; then
        ok "OpenRouter API key valid"
    else
        warn "OpenRouter API returned $STATUS"
    fi
else
    warn "OPENROUTER_API_KEY not set — AI features disabled"
fi

if [ "$CHECKS_PASSED" = false ]; then
    echo ""
    fail "Required dependencies missing. Fix above errors and retry."
    exit 1
fi

# ── Install frontend deps if needed ──
if [ ! -d "$ROOT/frontend/node_modules" ]; then
    echo ""
    echo "Installing frontend dependencies..."
    (cd "$ROOT/frontend" && npm install)
fi

# ── Start servers ──
echo ""
echo "Starting servers:"

# Backend
(cd "$ROOT" && go run ./cmd/server) > "$ROOT/.dev-backend.log" 2>&1 &
BACKEND_PID=$!
echo "$BACKEND_PID backend" > "$PIDFILE"
ok "Backend starting (PID $BACKEND_PID) → http://localhost:${PORT:-8080}"

# Frontend
(cd "$ROOT/frontend" && npm run dev -- --port "${FRONTEND_PORT:-3000}") > "$ROOT/.dev-frontend.log" 2>&1 &
FRONTEND_PID=$!
echo "$FRONTEND_PID frontend" >> "$PIDFILE"
ok "Frontend starting (PID $FRONTEND_PID) → http://localhost:${FRONTEND_PORT:-3000}"

# Wait for backend to be ready
echo ""
echo -n "Waiting for backend health..."
for i in $(seq 1 30); do
    if curl -s "http://localhost:${PORT:-8080}/health" 2>/dev/null | grep -q "ok"; then
        echo -e " ${GREEN}ready${NC}"
        break
    fi
    if ! kill -0 "$BACKEND_PID" 2>/dev/null; then
        echo -e " ${RED}crashed${NC}"
        echo "Backend log:"
        tail -20 "$ROOT/.dev-backend.log"
        rm -f "$PIDFILE"
        exit 1
    fi
    sleep 1
    echo -n "."
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e " ${GREEN}All servers running${NC}"
echo "  Backend:  http://localhost:${PORT:-8080}"
echo "  Frontend: http://localhost:${FRONTEND_PORT:-3000}"
echo "  Health:   http://localhost:${PORT:-8080}/health"
echo "  Ready:    http://localhost:${PORT:-8080}/ready"
echo ""
echo "  Logs: tail -f .dev-backend.log .dev-frontend.log"
echo "  Stop: ./scripts/stop.sh"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
