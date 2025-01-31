package email

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type EmailRepository interface {
	SaveVerificationCode(ctx context.Context, userID int64, code string, expiresAt time.Time) error
	GetVerificationCode(ctx context.Context, userID int64) (*VerificationCode, error)
	DeleteVerificationCode(ctx context.Context, userID int64) error
}

type VerificationCode struct {
	Code      string
	ExpiresAt time.Time
}

type emailRepositoryImpl struct {
	db *sql.DB
}

type VerificationRepositoryImpl struct {
	db *sql.DB
}

func NewEmailRepository(db *sql.DB) EmailRepository {
	return &emailRepositoryImpl{db: db}
}

func (r *emailRepositoryImpl) SaveVerificationCode(ctx context.Context, userID int64, code string, expiresAt time.Time) error {
	query := `
		INSERT INTO email_verifications (user_id, code, expires_at)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.ExecContext(ctx, query, userID, code, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to save verification code: %w", err)
	}
	return nil
}

func (r *emailRepositoryImpl) GetVerificationCode(ctx context.Context, userID int64) (*VerificationCode, error) {
	query := `
		SELECT code, expires_at
		FROM email_verifications
		WHERE user_id = $1
	`
	code := &VerificationCode{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&code.Code, &code.ExpiresAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get verification code: %w", err)
	}
	return code, nil
}

func (r *emailRepositoryImpl) DeleteVerificationCode(ctx context.Context, userID int64) error {
	query := `
		DELETE FROM email_verifications
		WHERE user_id = $1
	`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete verification code: %w", err)
	}
	return nil
}
