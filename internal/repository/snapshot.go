package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/chessgoddess/chessgoddess/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SnapshotRepository struct {
	pool *pgxpool.Pool
}

func NewSnapshotRepository(pool *pgxpool.Pool) *SnapshotRepository {
	return &SnapshotRepository{pool: pool}
}

func generateShareToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (r *SnapshotRepository) Create(ctx context.Context, sessionID, userID string, data map[string]interface{}) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot data: %w", err)
	}

	shareToken, err := generateShareToken()
	if err != nil {
		return fmt.Errorf("failed to generate share token: %w", err)
	}

	query := `
		INSERT INTO snapshots (session_id, user_id, data, share_token, is_public)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	snapshot := &model.Snapshot{}
	err = r.pool.QueryRow(ctx, query, sessionID, userID, dataJSON, shareToken, false).
		Scan(&snapshot.ID, &snapshot.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create snapshot: %w", err)
	}

	snapshot.SessionID = sessionID
	snapshot.UserID = userID
	snapshot.ShareToken = shareToken

	return nil
}

func (r *SnapshotRepository) GetByShareToken(ctx context.Context, shareToken string) (*model.Snapshot, error) {
	query := `
		SELECT id, session_id, user_id, data, share_token, is_public, created_at
		FROM snapshots
		WHERE share_token = $1
	`

	snapshot := &model.Snapshot{}
	var dataJSON []byte

	err := r.pool.QueryRow(ctx, query, shareToken).
		Scan(&snapshot.ID, &snapshot.SessionID, &snapshot.UserID, &dataJSON, &snapshot.ShareToken, &snapshot.IsPublic, &snapshot.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}

	if err := json.Unmarshal(dataJSON, &snapshot.Data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshot data: %w", err)
	}

	return snapshot, nil
}

func (r *SnapshotRepository) GetByID(ctx context.Context, id string) (*model.Snapshot, error) {
	query := `
		SELECT id, session_id, user_id, data, share_token, is_public, created_at
		FROM snapshots
		WHERE id = $1
	`

	snapshot := &model.Snapshot{}
	var dataJSON []byte

	err := r.pool.QueryRow(ctx, query, id).
		Scan(&snapshot.ID, &snapshot.SessionID, &snapshot.UserID, &dataJSON, &snapshot.ShareToken, &snapshot.IsPublic, &snapshot.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}

	if err := json.Unmarshal(dataJSON, &snapshot.Data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshot data: %w", err)
	}

	return snapshot, nil
}

func (r *SnapshotRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]model.Snapshot, error) {
	query := `
		SELECT id, session_id, user_id, data, share_token, is_public, created_at
		FROM snapshots
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list snapshots: %w", err)
	}
	defer rows.Close()

	var snapshots []model.Snapshot
	for rows.Next() {
		var snapshot model.Snapshot
		var dataJSON []byte

		err := rows.Scan(&snapshot.ID, &snapshot.SessionID, &snapshot.UserID, &dataJSON, &snapshot.ShareToken, &snapshot.IsPublic, &snapshot.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan snapshot: %w", err)
		}

		if err := json.Unmarshal(dataJSON, &snapshot.Data); err != nil {
			continue
		}

		snapshots = append(snapshots, snapshot)
	}

	return snapshots, nil
}
