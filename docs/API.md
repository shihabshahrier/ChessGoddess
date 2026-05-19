# ChessLens API Reference

Base URL: `/api/v1`

## Public Endpoints

### Health Check
```
GET /health
```
Returns `200 OK` with `{"status": "ok", "service": "chesslens"}`.

### Readiness Check
```
GET /ready
```
Pings PostgreSQL and Redis. Returns `200` if both healthy, `503` if either is down.

Response (healthy):
```json
{"status": "ready"}
```

Response (unhealthy):
```json
{"status": "not ready", "reason": "database unavailable"}
```

---

## Auth

### Get Google OAuth URL
```
GET /api/v1/auth/google/url
```
Returns the Google OAuth2 authorization URL. Sets `oauth_state` cookie for CSRF protection.

Response:
```json
{"url": "https://accounts.google.com/o/oauth2/auth?..."}
```

### Google OAuth Callback
```
GET /api/v1/auth/google/callback?state=...&code=...
```
Validates state cookie, exchanges authorization code, fetches Google user info, upserts user in DB, sets `auth_token` JWT cookie, and redirects to `/dashboard`.

---

## Games (JWT Required)

### Upload Game
```
POST /api/v1/games/upload
Content-Type: application/json
Authorization: Bearer <token>
```
Body (max 1MB):
```json
{"pgn": "1. e4 e5 2. Nf3 Nc6 ..."}
```
Response `201`:
```json
{"message": "game uploaded successfully", "game_id": "uuid"}
```

### List Games
```
GET /api/v1/games
Authorization: Bearer <token>
```
Response:
```json
{"games": [{"id": "...", "pgn": "...", "white_player": "...", ...}]}
```

### Get Game
```
GET /api/v1/games/:id
Authorization: Bearer <token>
```

### Delete Game
```
DELETE /api/v1/games/:id
Authorization: Bearer <token>
```

---

## Analysis (JWT Required)

### Create Analysis Session
```
POST /api/v1/analysis
Authorization: Bearer <token>
```
Body:
```json
{"game_id": "uuid", "depth": 20}
```
Response `201`:
```json
{"message": "analysis session created", "session_id": "uuid", "status": "pending"}
```

### Get Analysis Session
```
GET /api/v1/analysis/:id
Authorization: Bearer <token>
```

---

## Snapshots

### Get Public Snapshot (No Auth)
```
GET /api/v1/snapshots/:token
```

### Create Snapshot (JWT Required)
```
POST /api/v1/snapshots?session_id=uuid
Authorization: Bearer <token>
```

### List User Snapshots (JWT Required)
```
GET /api/v1/snapshots
Authorization: Bearer <token>
```

---

## AI Explanations (JWT Required)

### Explain Move
```
POST /api/v1/ai/explain
Authorization: Bearer <token>
```
Body:
```json
{"move_id": "uuid", "session_id": "uuid"}
```

### Explain Blunder
```
POST /api/v1/ai/explain-blunder
Authorization: Bearer <token>
```
Body:
```json
{"move_id": "uuid", "session_id": "uuid"}
```

### Get Cached Explanation
```
GET /api/v1/ai/explanation/:move_id
Authorization: Bearer <token>
```

---

## Vision (JWT Required)

### Image to FEN (File Upload)
```
POST /api/v1/vision/image-to-fen
Authorization: Bearer <token>
Content-Type: multipart/form-data
```
Form field: `image` (max 10MB, PNG/JPG)

### Image to FEN (URL)
```
POST /api/v1/vision/image-to-fen-url
Authorization: Bearer <token>
```
Body:
```json
{"url": "https://example.com/board.png"}
```

---

## Error Format

All errors return:
```json
{"error": "human-readable error message"}
```

## Rate Limiting

All endpoints are rate-limited to 10 requests/second per IP with burst capacity of 30. Exceeding returns `429 Too Many Requests`.

## Authentication

JWT tokens are issued as `auth_token` HttpOnly cookies during OAuth callback. Alternatively, pass `Authorization: Bearer <token>` header. Tokens expire after 7 days.
