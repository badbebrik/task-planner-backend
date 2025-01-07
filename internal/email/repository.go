package email

import (
	"database/sql"
	"fmt"
	"time"
)

type EmailRepository interface {
	SaveVerificationCode(userID int, code string, expiresAt time.Time) error
}

type EmailRepositoryImpl struct {
	db *sql.DB
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
