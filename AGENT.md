# 🧠 ChessGoddess - Agent Knowledge Graph & Project Plan

## 📊 Project Overview

**Product Name:** ChessGoddess  
**Public Brand:** ChessLens  
**Core Idea:** A cinematic chess analysis studio that turns raw engine output into readable insight, visual tension, and beautiful review experiences.

**Tech Stack:**
- Frontend: React
- Backend: Go (Gin)
- Database: Postgres
- Cache/Queue: Redis
- Storage: Cloudflare R2
- Engine: Stockfish
- AI: OpenRouter (LLM explanations)
- Auth: Google SSO
- Deployment: Cloudflare Pages (frontend), Cloud Run/ECS (backend)

---

## 🔗 Knowledge Graph

```
ChessGoddess (Project)
├── Frontend (React)
│   ├── UI Components
│   │   ├── ChessBoard
│   │   ├── EvalBar
│   │   ├── MoveList
│   │   ├── AnalysisPanel
│   │   ├── ReviewUI
│   │   └── ShareCard
│   ├── Pages
│   │   ├── Home
│   │   ├── Upload
│   │   ├── Analysis
│   │   ├── Review
│   │   └── Share (/s/:snapshot_id)
│   ├── State Management
│   │   ├── Game State
│   │   ├── Analysis State
│   │   └── UI State
│   └── Motion System
│       ├── Piece Glide
│       ├── Board Shake
│       ├── Eval Bar Animation
│       └── Scroll Timeline
│
├── Backend (Go/Gin)
│   ├── cmd/
│   │   └── server/
│   └── internal/
│       ├── auth/          → Google SSO, JWT
│       ├── analysis/      → Analysis orchestration
│       ├── engine/        → Stockfish worker management
│       ├── review/        → Review session handling
│       ├── snapshot/      → Immutable snapshot creation
│       ├── ai/            → OpenRouter LLM integration
│       ├── storage/       → R2 upload handling
│       ├── middleware/    → Auth, logging, CORS
│       └── websocket/     → Real-time streaming
│
├── Database (Postgres)
│   ├── users              → Google SSO identity
│   ├── games              → PGN/FEN source, metadata
│   ├── analysis_sessions  → Review instance state
│   ├── moves              → FEN, SAN, eval, classification
│   ├── snapshots          → Immutable JSON analysis
│   ├── ai_explanations    → Cached LLM output
│   └── uploads            → R2 reference metadata
│
├── Infrastructure
│   ├── Redis              → Queue/Cache
│   ├── R2                 → Screenshots, board images
│   ├── Stockfish Workers  → Engine evaluation
│   └── OpenRouter         → AI explanations
│
└── CI/CD
    ├── Frontend           → Lint, typecheck, build, deploy
    └── Backend            → Test, build docker, deploy
```

---

## 📋 Implementation Phases

### Phase 1: Foundation & MVP
**Goal:** Core infrastructure, basic analysis, Google auth

| Module | Description | Status |
|--------|-------------|--------|
| 1.1 | Project scaffolding (Go + React) | ✅ Complete |
| 1.2 | Database schema & migrations | ✅ Complete |
| 1.3 | Google SSO auth flow | ✅ Complete |
| 1.4 | PGN/FEN upload & parsing | ✅ Complete |
| 1.5 | Stockfish integration (local) | ✅ Complete |
| 1.6 | Basic analysis API | ✅ Complete |
| 1.7 | React chess board component | ✅ Complete |
| 1.8 | Basic analysis UI | ✅ Complete |
| 1.9 | Docker compose for local dev | ✅ Complete |

### Phase 2: Analysis & Review System
**Goal:** Full analysis pipeline, review UI, snapshots

| Module | Description | Status |
|--------|-------------|--------|
| 2.1 | Redis queue for analysis jobs | ✅ Complete |
| 2.2 | Worker scaling architecture | ✅ Complete |
| 2.3 | Move classification (blunder/mistake/etc) | ✅ Complete |
| 2.4 | Immutable snapshot creation | ✅ Complete |
| 2.5 | Review UI with timeline | ✅ Complete |
| 2.6 | Eval bar with spring physics | ✅ Complete |
| 2.7 | Piece glide animations | ✅ Complete |
| 2.8 | Scroll-to-scrub timeline | ✅ Complete |

### Phase 3: AI Layer & Sharing
**Goal:** AI explanations, sharing ecosystem, polish

| Module | Description | Status |
|--------|-------------|--------|
| 3.1 | OpenRouter integration | ✅ Complete |
| 3.2 | AI explanation caching | ✅ Complete |
| 3.3 | "Why this move is bad" feature | ✅ Complete |
| 3.4 | Share links (/s/:snapshot_id) | ✅ Complete |
| 3.5 | Open Graph preview | ✅ Complete |
| 3.6 | R2 storage for screenshots | ✅ Complete |
| 3.7 | Image-to-FEN (optional) | ✅ Complete |

### Phase 4: Polish & Production
**Goal:** Testing, CI/CD, deployment, performance

| Module | Description | Status |
|--------|-------------|--------|
| 4.1 | Frontend tests (Playwright + unit) | ✅ Complete |
| 4.2 | Backend tests (Go tests) | ✅ Complete |
| 4.3 | Snapshot integrity tests | ✅ Complete |
| 4.4 | GitHub Actions CI/CD | ✅ Complete |
| 4.5 | Cloudflare Pages deploy | ⬜ Config Ready |
| 4.6 | Cloud Run/ECS deploy | ⬜ Config Ready |
| 4.7 | Performance optimization | ⬜ Config Ready |
| 4.8 | Dark chess hall aesthetic | ✅ Complete |

---

## 🔄 Git Commit Strategy

### Branch Strategy
- `main` - Production-ready code
- `develop` - Integration branch
- `feature/*` - Feature branches

### Commit Message Convention
```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Types:**
- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation
- `style` - Formatting
- `refactor` - Code restructuring
- `test` - Tests
- `chore` - Maintenance

### Planned Commit Sequence (Phase 1)

1. `chore: initialize project structure`
   - Go module setup
   - React app scaffolding
   - Basic directory structure

2. `feat(db): create database schema and migrations`
   - Postgres schema
   - Migration files
   - Database connection

3. `feat(auth): implement Google SSO authentication`
   - OAuth flow
   - JWT generation
   - User model

4. `feat(upload): add PGN/FEN upload and parsing`
   - File upload endpoint
   - PGN parsing
   - Game model

5. `feat(engine): integrate Stockfish for analysis`
   - Stockfish process management
   - Evaluation parsing
   - Worker interface

6. `feat(api): create analysis API endpoints`
   - Analysis session creation
   - Move-by-move analysis
   - Results storage

7. `feat(ui): build chess board component`
   - Board rendering
   - Piece display
   - Move highlighting

8. `feat(ui): create analysis interface`
   - Upload page
   - Analysis display
   - Move list

9. `chore(devops): add Docker Compose for local development`
   - Postgres service
   - Redis service
   - Backend service
   - Frontend service

---

## 📊 Progress Tracking

### Overall Progress
```
Phase 1: ✅✅✅✅✅✅✅✅✅ 100%
Phase 2: ✅✅✅✅✅✅✅✅ 100%
Phase 3: ✅✅✅✅✅✅✅ 100%
Phase 4: ✅✅✅✅⬜⬜⬜✅ 63%
```

### Total: 29/32 modules complete (91%)

---

## 🎯 Next Steps

1. **Deploy:** Configure Cloudflare Pages for frontend
2. **Deploy:** Set up Cloud Run / ECS for backend
3. **Deploy:** Configure environment variables and secrets
4. **Monitor:** Set up logging and alerting

---

## 📝 Notes

- All snapshots are immutable once created
- Cache all LLM calls aggressively
- Stockfish runs as worker processes
- Frontend uses dark chess hall aesthetic
- Motion system is critical to product identity
- Share links are read-only and heavily cached
- Docker NOT installed on this machine - use Homebrew or cloud services
- 23 Go unit tests passing, integration tests skipped (no DB/Redis)
- Playwright e2e tests written but not executed

---

*Last Updated: 2026-05-19*
*Status: 29/32 modules complete (91%) - Core features done, Docker not installed, 23 Go tests passing*
