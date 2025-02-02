package auth

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type RefreshToken struct {
	ID        int64
	UserID    int64
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type TokenRepository interface {
	SaveRefreshToken(ctx context.Context, userID int64, token string, expiresAt time.Time) error
	GetRefreshToken(ctx context.Context, token string) (*RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, token string) error
}

type tokenRepositoryImpl struct {
	db *sql.DB
}

func NewTokenRepository(db *sql.DB) TokenRepository {
	return &tokenRepositoryImpl{db: db}
}

func (r *tokenRepositoryImpl) SaveRefreshToken(ctx context.Context, userID int64, token string, expiresAt time.Time) error {
	query := `
        INSERT INTO refresh_tokens (user_id, token, expires_at, created_at)
        VALUES ($1, $2, $3, NOW())
    `
	_, err := r.db.ExecContext(ctx, query, userID, token, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to save refresh token: %w", err)
	}
	return nil
}

func (r *tokenRepositoryImpl) GetRefreshToken(ctx context.Context, token string) (*RefreshToken, error) {
	query := `
        SELECT id, user_id, token, expires_at, created_at
        FROM refresh_tokens
        WHERE token = $1
        `
	var rt RefreshToken
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&rt.ID,
		&rt.UserID,
		&rt.Token,
		&rt.ExpiresAt,
		&rt.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}
	return &rt, nil
}

func (r *tokenRepositoryImpl) DeleteRefreshToken(ctx context.Context, token string) error {
	query := `DELETE FROM refresh_tokens WHERE token = $1`
	_, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}
	return nil
}
