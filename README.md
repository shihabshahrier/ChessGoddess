

# 🧠 Product Identity: ChessGoddess

## Core idea

> A cinematic chess analysis studio that turns raw engine output into readable insight, visual tension, and beautiful review experiences.

Not:

* just analysis
* not just engine output

More like:

* “you can *see* thinking”

---

# 🧩 System Naming (important for architecture)

You still keep your internal module name clean and boring where it should be.

## Public brand:

* ChessLens

## Internal system components:

* Lens Core (analysis engine layer)
* Lens Review System
* Lens AI Coach
* Lens Snapshot Engine

And yes, Stockfish still lives underneath everything like a slightly annoyed genius doing all the work.

---

# 🏗️ High-Level Architecture (updated for ChessLens)

```txt id="chesslens_arch"
Frontend (React)
   ↓
Gin API (ChessLens Core API)
   ↓
-----------------------------
| Services Layer           |
| - Auth (Google SSO)      |
| - Analysis Service       |
| - Review Service         |
| - AI Explanation Layer   |
| - Snapshot Service       |
-----------------------------
   ↓
Postgres (source of truth)
Redis (queue/cache)
R2 (uploads/screenshots)
Stockfish Workers
OpenRouter (LLM explanations)
```

---

# 🎯 Core Product Loop

This is your real system behavior:

```txt id="loop1"
Upload / Import Game
      ↓
Create Analysis Session
      ↓
Run Stockfish evaluation
      ↓
Generate move-by-move insights
      ↓
Store immutable snapshot
      ↓
Optional AI explanation layer
      ↓
Beautiful animated review UI
      ↓
Share link
```

---

# 🧬 Data Model (refined)

## users

* identity layer (Google SSO)
* minimal profile

## games

* PGN / FEN source
* metadata (opening, result, time control)

## analysis_sessions

* single “review instance”
* engine config used
* depth settings
* timestamped state

## moves

* fen
* san
* evaluation
* best move
* classification (blunder/mistake/etc)

## snapshots (IMPORTANT)

* immutable JSON of full analysis
* shared via link
* never edited again

## ai_explanations

* move-level reasoning
* cached LLM output

## uploads

* screenshots stored in R2

---

# 🔥 Snapshot Philosophy (this is key)

You said:

> immutable review snapshots

Good. This is *product-grade thinking*.

So each analysis becomes:

> a frozen cinematic artifact

Not a mutable dashboard.

That means:

* share link = exact preserved analysis
* no drift
* no recomputation inconsistencies
* feels like “published work”

This is a huge UX win.

---

# 🎨 UI / Motion Direction

## Visual identity

* dark chess hall aesthetic
* walnut wood texture overlays
* soft gold accents
* deep charcoal background
* subtle vignette edges

---

## Motion system

### Eval bar behavior

* spring-based physics
* slow decay smoothing
* dramatic swing on blunders
* subtle “settling” animation after engine stabilizes

---

### Move transitions

* piece glide (not teleport)
* board micro-shake on blunders (very subtle)
* highlight trail for last move

---

### Review flow

* scroll = timeline scrub
* hovering move = instant board jump
* smooth interpolation between positions

---

# 🧠 AI Layer (OpenRouter)

You’ll use it for:

* “why this move is bad”
* positional explanations
* pattern recognition
* human-readable summaries

## Key design rule:

Cache everything.

Because:

* same positions repeat a lot
* LLM calls are the only real cost multiplier in your system

---

# ⚙️ Backend Structure (Gin)

```txt id="backend"
cmd/
internal/
  auth/
  analysis/
  engine/
  review/
  snapshot/
  ai/
  storage/
  middleware/
  websocket/
configs/
migrations/
```

---

# ⚡ Engine Layer (Stockfish workers)

Model:

* each analysis request becomes a job
* Redis queue handles execution
* worker runs Stockfish process
* results streamed back or stored

This gives you:

* scalability later
* non-blocking API
* deterministic results

---

# 🔐 Auth Flow (Google only)

Flow:

```txt id="auth"
React → Google OAuth → JWT → Gin verify → user session
```

Keep it dead simple.

No password system.
No email verification drama.
No “forgot password” edge-case hell.

---

# 📦 Storage Strategy

## R2 usage:

* screenshots
* board images
* shared previews
* thumbnails for review cards

Everything else:

* Postgres

---

# 🔗 Sharing System

Each snapshot:

```txt id="share"
/s/:snapshot_id
```

Rules:

* read-only
* immutable
* cached heavily
* open graph preview enabled

---

# 🧪 Testing Strategy

## Frontend

* Playwright (critical flows)
* unit tests for board logic

## Backend

* Go tests for analysis engine
* snapshot integrity tests
* API contract tests

## Engine tests

* FEN consistency validation
* evaluation stability checks

---

# 🚀 CI/CD (GitHub Actions)

## Frontend

* lint
* typecheck
* build
* deploy (Cloudflare Pages)

## Backend

* test
* build docker
* deploy (Cloud Run / ECS)

---

# 🧭 Deployment Plan

## Phase 1 (MVP)

* React frontend
* Gin backend
* local Stockfish
* Postgres + R2
* Google login
* basic analysis

## Phase 2

* Redis queue
* worker scaling
* AI explanations
* image-to-FEN

## Phase 3

* public sharing ecosystem
* study boards
* collaboration

---

# 🧨 Product positioning (important)

ChessLens is not:

* a chess engine UI
* a tool

It is:

> a visual reasoning layer over chess

That distinction matters because it defines:

* UX polish requirements
* motion design importance
* explanation layer priority

---

# Final state of the universe

You now have:

* Name: ChessGoddess ✔
* Architecture: defined ✔
* Tech stack: locked ✔
* infra model: realistic ✔
* product philosophy: clear ✔
* differentiation: UX + AI + motion ✔

