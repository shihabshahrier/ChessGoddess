#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PIDFILE="$ROOT/.dev.pids"

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " ChessGoddess — Stopping Dev Servers"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ ! -f "$PIDFILE" ]; then
    echo -e "  ${RED}✗${NC} No .dev.pids found — nothing to stop"

    # Kill any strays
    STRAYS=$(pgrep -f "go run ./cmd/server" 2>/dev/null || true)
    if [ -n "$STRAYS" ]; then
        echo "  Found stray Go processes: $STRAYS"
        kill $STRAYS 2>/dev/null || true
        echo -e "  ${GREEN}✓${NC} Killed stray processes"
    fi

    VITE_STRAYS=$(pgrep -f "vite.*--port" 2>/dev/null || true)
    if [ -n "$VITE_STRAYS" ]; then
        echo "  Found stray Vite processes: $VITE_STRAYS"
        kill $VITE_STRAYS 2>/dev/null || true
        echo -e "  ${GREEN}✓${NC} Killed stray Vite processes"
    fi

    exit 0
fi

while IFS=' ' read -r pid name; do
    if kill -0 "$pid" 2>/dev/null; then
        kill "$pid" 2>/dev/null || true
        # Wait up to 5s for graceful shutdown
        for i in $(seq 1 10); do
            if ! kill -0 "$pid" 2>/dev/null; then
                break
            fi
            sleep 0.5
        done
        # Force kill if still alive
        if kill -0 "$pid" 2>/dev/null; then
            kill -9 "$pid" 2>/dev/null || true
        fi
        echo -e "  ${GREEN}✓${NC} Stopped $name (PID $pid)"
    else
        echo -e "  ${GREEN}✓${NC} $name (PID $pid) already stopped"
    fi
done < "$PIDFILE"

rm -f "$PIDFILE"
rm -f "$ROOT/.dev-backend.log" "$ROOT/.dev-frontend.log"

echo ""
echo -e "  ${GREEN}All servers stopped${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
