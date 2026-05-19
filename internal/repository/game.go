package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/chessgoddess/chesslens/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GameRepository struct {
	pool *pgxpool.Pool
}

func NewGameRepository(pool *pgxpool.Pool) *GameRepository {
	return &GameRepository{pool: pool}
}

func (r *GameRepository) Create(ctx context.Context, game *model.Game) error {
	query := `
		INSERT INTO games (user_id, pgn, fen, white_player, black_player, result, opening, time_control, event, date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`
	err := r.pool.QueryRow(ctx, query,
		game.UserID, game.PGN, game.FEN, game.WhitePlayer, game.BlackPlayer,
		game.Result, game.Opening, game.TimeControl, game.Event, game.Date,
	).Scan(&game.ID, &game.CreatedAt, &game.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create game: %w", err)
	}
	return nil
}

func (r *GameRepository) GetByID(ctx context.Context, id string) (*model.Game, error) {
	query := `
		SELECT id, user_id, pgn, fen, white_player, black_player, result, opening, time_control, event, date, created_at, updated_at
		FROM games
		WHERE id = $1
	`
	game := &model.Game{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&game.ID, &game.UserID, &game.PGN, &game.FEN, &game.WhitePlayer, &game.BlackPlayer,
		&game.Result, &game.Opening, &game.TimeControl, &game.Event, &game.Date,
		&game.CreatedAt, &game.UpdatedAt,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}
	return game, nil
}

func (r *GameRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]model.Game, error) {
	query := `
		SELECT id, user_id, pgn, fen, white_player, black_player, result, opening, time_control, event, date, created_at, updated_at
		FROM games
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list games: %w", err)
	}
	defer rows.Close()
	
	var games []model.Game
	for rows.Next() {
		var game model.Game
		err := rows.Scan(
			&game.ID, &game.UserID, &game.PGN, &game.FEN, &game.WhitePlayer, &game.BlackPlayer,
			&game.Result, &game.Opening, &game.TimeControl, &game.Event, &game.Date,
			&game.CreatedAt, &game.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}
		games = append(games, game)
	}
	
	return games, nil
}

func (r *GameRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM games WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete game: %w", err)
	}
	return nil
}

type AnalysisSessionRepository struct {
	pool *pgxpool.Pool
}

func NewAnalysisSessionRepository(pool *pgxpool.Pool) *AnalysisSessionRepository {
	return &AnalysisSessionRepository{pool: pool}
}

func (r *AnalysisSessionRepository) Create(ctx context.Context, session *model.AnalysisSession) error {
	query := `
		INSERT INTO analysis_sessions (game_id, user_id, engine_config, depth, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	err := r.pool.QueryRow(ctx, query,
		session.GameID, session.UserID, session.EngineConfig, session.Depth, session.Status,
	).Scan(&session.ID, &session.CreatedAt, &session.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create analysis session: %w", err)
	}
	return nil
}

func (r *AnalysisSessionRepository) GetByID(ctx context.Context, id string) (*model.AnalysisSession, error) {
	query := `
		SELECT id, game_id, user_id, engine_config, depth, status, started_at, completed_at, created_at, updated_at
		FROM analysis_sessions
		WHERE id = $1
	`
	session := &model.AnalysisSession{}
	var engineConfig []byte
	
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&session.ID, &session.GameID, &session.UserID, &engineConfig, &session.Depth,
		&session.Status, &session.StartedAt, &session.CompletedAt, &session.CreatedAt, &session.UpdatedAt,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get analysis session: %w", err)
	}
	
	session.EngineConfig = string(engineConfig)
	return session, nil
}

func (r *AnalysisSessionRepository) UpdateStatus(ctx context.Context, id, status string) error {
	query := `UPDATE analysis_sessions SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update session status: %w", err)
	}
	return nil
}
