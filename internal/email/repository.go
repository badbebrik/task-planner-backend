package email

import (
	"database/sql"
	"fmt"
	"time"
)

type EmailRepository interface {
	SaveVerificationCode(userID int, code string, expiresAt time.Time) error
	VerificationRepository
}

type VerificationRepository interface {
	GetVerificationCode(userID int64) (*VerificationCode, error)
	DeleteVerificationCode(userID int64) error
}

type VerificationCode struct {
	Code      string
	ExpiresAt time.Time
}

type EmailRepositoryImpl struct {
	db *sql.DB
}

type VerificationRepositoryImpl struct {
	db *sql.DB
}

func NewVerificationRepository(db *sql.DB) *VerificationRepositoryImpl {
	return &VerificationRepositoryImpl{db: db}
}

func NewEmailRepository(db *sql.DB) *EmailRepositoryImpl {
	return &EmailRepositoryImpl{db: db}
}

func (r *EmailRepositoryImpl) SaveVerificationCode(userID int, code string, expiresAt time.Time) error {
	query := `
		INSERT INTO email_verifications (user_id, code, expires_at)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.Exec(query, userID, code, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to save verification code: %w", err)
	}
	return nil
}

func (r *EmailRepositoryImpl) GetVerificationCode(userID int64) (*VerificationCode, error) {
	query := `
		SELECT code, expires_at
		FROM email_verifications
		WHERE user_id = $1
	`
	code := &VerificationCode{}
	err := r.db.QueryRow(query, userID).Scan(&code.Code, &code.ExpiresAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get verification code: %w", err)
	}
	return code, nil
}

func (r *EmailRepositoryImpl) DeleteVerificationCode(userID int64) error {
	query := `
		DELETE FROM email_verifications
		WHERE user_id = $1
	`
	_, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete verification code: %w", err)
	}
	return nil
}
