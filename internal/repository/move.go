package repository

import (
	"context"
	"fmt"

	"github.com/chessgoddess/chessgoddess/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MoveRepository struct {
	pool *pgxpool.Pool
}

func NewMoveRepository(pool *pgxpool.Pool) *MoveRepository {
	return &MoveRepository{pool: pool}
}

const moveColumns = `id, session_id, move_number, fen, san, evaluation, best_move, classification, depth,
	eval_before, eval_after, cp_loss, accuracy, best_line, created_at`

func scanMove(row interface {
	Scan(dest ...any) error
}, move *model.Move) error {
	return row.Scan(
		&move.ID, &move.SessionID, &move.MoveNumber, &move.FEN, &move.SAN,
		&move.Evaluation, &move.BestMove, &move.Classification, &move.Depth,
		&move.EvalBefore, &move.EvalAfter, &move.CPLoss, &move.Accuracy, &move.BestLine,
		&move.CreatedAt,
	)
}

func (r *MoveRepository) Create(ctx context.Context, move *model.Move) error {
	query := `
		INSERT INTO moves (
			session_id, move_number, fen, san, evaluation, best_move, classification, depth,
			eval_before, eval_after, cp_loss, accuracy, best_line
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at
	`
	err := r.pool.QueryRow(ctx, query,
		move.SessionID, move.MoveNumber, move.FEN, move.SAN,
		move.Evaluation, move.BestMove, move.Classification, move.Depth,
		move.EvalBefore, move.EvalAfter, move.CPLoss, move.Accuracy, move.BestLine,
	).Scan(&move.ID, &move.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create move: %w", err)
	}
	return nil
}

func (r *MoveRepository) ListBySessionID(ctx context.Context, sessionID string) ([]model.Move, error) {
	query := `SELECT ` + moveColumns + `
		FROM moves
		WHERE session_id = $1
		ORDER BY move_number ASC`
	rows, err := r.pool.Query(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list moves: %w", err)
	}
	defer rows.Close()

	var moves []model.Move
	for rows.Next() {
		var move model.Move
		if err := scanMove(rows, &move); err != nil {
			return nil, fmt.Errorf("failed to scan move: %w", err)
		}
		moves = append(moves, move)
	}

	return moves, nil
}

func (r *MoveRepository) GetByID(ctx context.Context, id string) (*model.Move, error) {
	query := `SELECT ` + moveColumns + ` FROM moves WHERE id = $1`
	move := &model.Move{}
	if err := scanMove(r.pool.QueryRow(ctx, query, id), move); err != nil {
		return nil, fmt.Errorf("failed to get move: %w", err)
	}
	return move, nil
}
