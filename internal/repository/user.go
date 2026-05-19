package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/chessgoddess/chesslens/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (email, name, avatar_url, google_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	err := r.pool.QueryRow(ctx, query, user.Email, user.Name, user.AvatarURL, user.GoogleID).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByGoogleID(ctx context.Context, googleID string) (*model.User, error) {
	query := `
		SELECT id, email, name, avatar_url, google_id, created_at, updated_at
		FROM users
		WHERE google_id = $1
	`
	user := &model.User{}
	err := r.pool.QueryRow(ctx, query, googleID).
		Scan(&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.GoogleID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by google ID: %w", err)
	}
	return user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	query := `
		SELECT id, email, name, avatar_url, google_id, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	user := &model.User{}
	err := r.pool.QueryRow(ctx, query, id).
		Scan(&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.GoogleID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users
		SET email = $1, name = $2, avatar_url = $3, updated_at = $4
		WHERE id = $5
	`
	_, err := r.pool.Exec(ctx, query, user.Email, user.Name, user.AvatarURL, time.Now(), user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}
