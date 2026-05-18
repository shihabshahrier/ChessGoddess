package repository

import (
	"context"
	"fmt"

	"github.com/chessgoddess/chesslens/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AIExplanationRepository struct {
	pool *pgxpool.Pool
}

func NewAIExplanationRepository(pool *pgxpool.Pool) *AIExplanationRepository {
	return &AIExplanationRepository{pool: pool}
}

func (r *AIExplanationRepository) Create(ctx context.Context, sessionID, moveID, fen, content, model string) error {
	query := `
		INSERT INTO ai_explanations (session_id, move_id, fen, content, model)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	
	var explanation models.AIExplanation
	err := r.pool.QueryRow(ctx, query, sessionID, moveID, fen, content, model).
		Scan(&explanation.ID, &explanation.CreatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create AI explanation: %w", err)
	}
	
	return nil
}

func (r *AIExplanationRepository) GetByMoveID(ctx context.Context, moveID string) (*models.AIExplanation, error) {
	query := `
		SELECT id, session_id, move_id, fen, content, model, created_at
		FROM ai_explanations
		WHERE move_id = $1
	`
	
	explanation := &models.AIExplanation{}
	err := r.pool.QueryRow(ctx, query, moveID).Scan(
		&explanation.ID, &explanation.SessionID, &explanation.MoveID,
		&explanation.FEN, &explanation.Content, &explanation.Model, &explanation.CreatedAt,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get AI explanation: %w", err)
	}
	
	return explanation, nil
}

func (r *AIExplanationRepository) GetByFEN(ctx context.Context, fen string) (*models.AIExplanation, error) {
	query := `
		SELECT id, session_id, move_id, fen, content, model, created_at
		FROM ai_explanations
		WHERE fen = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	
	explanation := &models.AIExplanation{}
	err := r.pool.QueryRow(ctx, query, fen).Scan(
		&explanation.ID, &explanation.SessionID, &explanation.MoveID,
		&explanation.FEN, &explanation.Content, &explanation.Model, &explanation.CreatedAt,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get AI explanation by FEN: %w", err)
	}
	
	return explanation, nil
}

func (r *AIExplanationRepository) ListBySessionID(ctx context.Context, sessionID string) ([]models.AIExplanation, error) {
	query := `
		SELECT id, session_id, move_id, fen, content, model, created_at
		FROM ai_explanations
		WHERE session_id = $1
		ORDER BY created_at ASC
	`
	
	rows, err := r.pool.Query(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list AI explanations: %w", err)
	}
	defer rows.Close()
	
	var explanations []models.AIExplanation
	for rows.Next() {
		var exp models.AIExplanation
		err := rows.Scan(
			&exp.ID, &exp.SessionID, &exp.MoveID,
			&exp.FEN, &exp.Content, &exp.Model, &exp.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan AI explanation: %w", err)
		}
		explanations = append(explanations, exp)
	}
	
	return explanations, nil
}
