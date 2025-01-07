package user

import (
	"database/sql"
	"fmt"
	"time"
)

type UserRepository interface {
	CreateUser(user *User) (int, error)
	UserExists(email string) (bool, error)
	GetUserByEmail(email string) (*User, error)
	MarkEmailAsVerified(userID int64) error
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

func (r *PGRepository) CreateUser(user *User) (int, error) {
	query := `
		INSERT INTO users (email, password_hash, name, is_email_verified, google_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	var userID int
	err := r.db.QueryRow(
		query,
		user.Email,
		user.PasswordHash,
		user.Name,
		user.IsEmailVerified,
		user.GoogleID,
		time.Now(),
		time.Now(),
	).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}
	return userID, nil
}

func (r *PGRepository) GetUserByEmail(email string) (*User, error) {
	query := `
		SELECT id, email, password_hash, name, is_email_verified, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	user := &User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name,
		&user.IsEmailVerified, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return user, nil
}

func (r *PGRepository) MarkEmailAsVerified(userID int64) error {
	query := `
		UPDATE users
		SET is_email_verified = TRUE, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to mark email as verified: %w", err)
	}
	return nil
}
