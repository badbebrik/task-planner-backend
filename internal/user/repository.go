package user

import (
	"database/sql"
	"fmt"
	"time"
)

type UserRepository interface {
	UserExists(email string) (bool, error)
	CreateUser(user *User) error
}

type PGRepository struct {
	db *sql.DB
}

func NewPGRepository(db *sql.DB) *PGRepository {
	return &PGRepository{db: db}
}

func (r *PGRepository) UserExists(email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)`
	err := r.db.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}
	return exists, nil
}

func (r *PGRepository) CreateUser(user *User) error {
	query := `
		INSERT INTO users (email, password_hash, name, is_email_verified, google_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(
		query,
		user.Email,
		user.PasswordHash,
		user.Name,
		user.IsEmailVerified,
		user.GoogleID,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}
