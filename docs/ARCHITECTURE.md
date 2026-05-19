# ChessLens Architecture

## System Overview

```
┌─────────────────────────────────────────────────────┐
│                   CLIENT (Browser)                  │
│  React + TypeScript + Tailwind + Zustand + Framer   │
│  Pages: Home → Upload → Analysis → Review → Share   │
└────────────────────┬────────────────────────────────┘
                     │ HTTPS / REST
┌────────────────────▼────────────────────────────────┐
│              BACKEND (Go / Gin)                     │
│  internal/api/     → HTTP handlers + routing        │
│  internal/auth/    → Google OAuth2 + JWT            │
│  internal/service/ → analysis, AI, vision           │
│  internal/engine/  → Stockfish process mgmt         │
│  internal/repository/ → DB access layer             │
│  internal/worker/  → Redis job queue + workers      │
│  internal/storage/ → Cloudflare R2                  │
└──────┬──────────────┬───────────────┬───────────────┘
       │              │               │
  PostgreSQL        Redis       Cloudflare R2
  (data store)   (queue/cache)  (file storage)
       │
  Stockfish      OpenRouter
  (local bin)    (LLM API)
```

## Package Responsibilities

### `cmd/server/main.go`
Entry point. Loads config, connects DB, starts HTTP server with graceful shutdown (SIGINT/SIGTERM, 30s drain). Uses `log/slog` with JSON output in production.

### `internal/api/`
HTTP layer. Gin router with middleware stack:
1. Recovery (panic → 500)
2. Logger (structured slog)
3. CORS (origin whitelist from `ALLOWED_ORIGINS`)
4. Rate limiter (10 req/s per IP, token bucket)

Routes split into handler structs: `AuthHandlers`, `GameHandlers`, `AIHandlers`, `VisionHandlers`, `SnapshotHandlers`. Each receives its dependencies via constructor.

### `internal/auth/`
Google OAuth2 flow + JWT (HS256). Handles state generation, code exchange, token signing/validation. Cookie management for auth tokens (HttpOnly, Secure in prod, SameSite=Lax, 7-day expiry).

### `internal/service/`
Business logic layer, no HTTP awareness.

- **AnalysisService** — Wraps Stockfish engine. Parses PGN, iterates moves, maintains FEN state via `notnil/chess`, runs engine evaluation per position, classifies moves (best/good/inaccuracy/mistake/blunder), persists results.
- **AIService** — OpenRouter LLM integration for move explanations. Redis caching for repeated queries. Supports explain-move and explain-blunder prompts.
- **VisionClient** — Image-to-FEN via OpenRouter vision models. Accepts file upload or URL.

### `internal/engine/`
Stockfish process management. Spawns stockfish binary, communicates via UCI protocol over stdin/stdout. Configurable depth, threads, hash size. Parses evaluation lines for centipawn score, best move, principal variation.

### `internal/repository/`
Data access layer. One struct per domain entity, takes `pgxpool.Pool`. No business logic — pure CRUD + queries.

### `internal/worker/`
Background job processing. Redis-backed queue (`worker.Queue`) with `Enqueue`/`Dequeue` for analysis and snapshot jobs. `worker.Worker` runs N goroutines polling the queue, dispatching to `AnalysisService`.

### `internal/db/`
PostgreSQL connection pool wrapper. `db.New()` creates pool from config, `db.Ping()` for health checks.

### `internal/config/`
Environment variable loading with defaults. Validates required vars, enforces JWT_SECRET strength in production, parses `ALLOWED_ORIGINS` comma list.

### `internal/model/`
Domain structs matching DB schema: User, Game, AnalysisSession, Move, Snapshot, AIExplanation, Upload.

### `internal/storage/`
Cloudflare R2 (S3-compatible) client for file uploads.

## Data Flow: Game Analysis

```
Upload PGN → Parse → Create Game row → Create AnalysisSession (pending)
    → Enqueue Redis job
    → Worker picks up job
    → Load Game PGN from DB
    → For each move:
        1. Parse SAN, advance FEN (notnil/chess)
        2. Send position to Stockfish
        3. Parse evaluation + best move
        4. Classify move quality
        5. Insert Move row
    → Update session status → completed
```

## Data Flow: OAuth Login

```
GET /auth/google/url → Generate state, set cookie, return Google URL
    → User consents on Google
GET /auth/google/callback?state=X&code=Y
    → Validate state cookie (CSRF)
    → Exchange code for OAuth token
    → Fetch Google userinfo (id, email, name, picture)
    → Upsert user in DB (by google_id)
    → Generate JWT (user_id, email, name, avatar)
    → Set auth_token cookie (HttpOnly, 7 days)
    → Redirect to /dashboard
```

## Security Measures

- **CORS**: Origin whitelist, no wildcards
- **JWT**: HS256, HttpOnly cookie, production secret validation
- **OAuth CSRF**: Random state in cookie, verified on callback
- **Rate limiting**: Per-IP token bucket (10/s, burst 30)
- **Upload limits**: PGN 1MB, images 10MB via `MaxBytesReader`
- **Cookie flags**: Secure (production), HttpOnly, SameSite=Lax
- **Error sanitization**: API error bodies not exposed in error messages
- **Context propagation**: `http.NewRequestWithContext` for external calls

## Database Schema

```
users
  ├── games (user_id FK)
  │   └── analysis_sessions (game_id FK, user_id FK)
  │       ├── moves (session_id FK)
  │       │   └── ai_explanations (session_id FK, move_id FK)
  │       └── snapshots (session_id FK, user_id FK)
  └── uploads (user_id FK)
```

## Deployment

- **Local**: `docker-compose.yml` — Go backend, React frontend, PostgreSQL, Redis
- **Production**: `docker-compose.prod.yml` — multi-stage builds, nginx for frontend
- **CI**: GitHub Actions — Go test/lint/build, frontend typecheck/lint/build, Docker image build
- **Deploy target**: Cloud Run, ECS, or Fly.io (not yet configured — see AGENT.md D6)
