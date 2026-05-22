// Package model defines domain types used across the application.
package model

import (
	"time"
)

type User struct {
	ID        string    `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Name      string    `json:"name" db:"name"`
	AvatarURL string    `json:"avatar_url" db:"avatar_url"`
	GoogleID  string    `json:"-" db:"google_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Game struct {
	ID          string    `json:"id" db:"id"`
	UserID      string    `json:"user_id" db:"user_id"`
	PGN         string    `json:"pgn" db:"pgn"`
	FEN         string    `json:"fen" db:"fen"`
	WhitePlayer string    `json:"white_player" db:"white_player"`
	BlackPlayer string    `json:"black_player" db:"black_player"`
	Result      string    `json:"result" db:"result"`
	Opening     string    `json:"opening" db:"opening"`
	TimeControl string    `json:"time_control" db:"time_control"`
	Event       string    `json:"event" db:"event"`
	Date        string    `json:"date" db:"date"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type AnalysisSession struct {
	ID            string     `json:"id" db:"id"`
	GameID        string     `json:"game_id" db:"game_id"`
	UserID        string     `json:"user_id" db:"user_id"`
	EngineConfig  string     `json:"engine_config" db:"engine_config"`
	Depth         int        `json:"depth" db:"depth"`
	Status        string     `json:"status" db:"status"` // pending, running, completed, failed
	AccuracyWhite float64    `json:"accuracy_white" db:"accuracy_white"`
	AccuracyBlack float64    `json:"accuracy_black" db:"accuracy_black"`
	StartedAt     time.Time  `json:"started_at" db:"started_at"`
	CompletedAt   *time.Time `json:"completed_at" db:"completed_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

type Move struct {
	ID             string    `json:"id" db:"id"`
	SessionID      string    `json:"session_id" db:"session_id"`
	MoveNumber     int       `json:"move_number" db:"move_number"`
	FEN            string    `json:"fen" db:"fen"`
	SAN            string    `json:"san" db:"san"`
	Evaluation     float64   `json:"evaluation" db:"evaluation"`         // eval after the move, pawns, white POV
	EvalBefore     float64   `json:"eval_before" db:"eval_before"`       // best eval available before the move, white POV
	EvalAfter      float64   `json:"eval_after" db:"eval_after"`         // eval after the move, white POV
	CPLoss         float64   `json:"cp_loss" db:"cp_loss"`               // centipawns lost vs best (>= 0)
	Accuracy       float64   `json:"accuracy" db:"accuracy"`             // 0-100 move accuracy
	BestMove       string    `json:"best_move" db:"best_move"`           // UCI long-algebraic
	BestLine       string    `json:"best_line" db:"best_line"`           // engine PV, space-separated UCI
	Classification string    `json:"classification" db:"classification"` // brilliant, great, best, excellent, good, book, inaccuracy, mistake, blunder
	Depth          int       `json:"depth" db:"depth"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

type Snapshot struct {
	ID         string                 `json:"id" db:"id"`
	SessionID  string                 `json:"session_id" db:"session_id"`
	UserID     string                 `json:"user_id" db:"user_id"`
	Data       map[string]interface{} `json:"data" db:"data"`
	ShareToken string                 `json:"share_token" db:"share_token"`
	IsPublic   bool                   `json:"is_public" db:"is_public"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
}

type AIExplanation struct {
	ID        string    `json:"id" db:"id"`
	SessionID string    `json:"session_id" db:"session_id"`
	MoveID    string    `json:"move_id" db:"move_id"`
	FEN       string    `json:"fen" db:"fen"`
	Content   string    `json:"content" db:"content"`
	Model     string    `json:"model" db:"model"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Upload struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	R2Key     string    `json:"r2_key" db:"r2_key"`
	URL       string    `json:"url" db:"url"`
	Type      string    `json:"type" db:"type"` // screenshot, board_image, thumbnail
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
