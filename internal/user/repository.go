package user

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Repository interface {
	GetByID(ctx context.Context, id int64) (*User, error)
	CreateUser(ctx context.Context, user *User) (int64, error)
	UserExists(ctx context.Context, email string) (bool, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	MarkEmailAsVerified(ctx context.Context, userID int64) error
	GetByGoogleID(ctx context.Context, googleID string) (*User, error)
	CreateWithGoogle(ctx context.Context, email, googleID string) (int64, error)
	LinkGoogleID(ctx context.Context, userID int64, googleID string) error
}

type PGRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *PGRepository {
	return &PGRepository{db: db}
}

func (r *PGRepository) UserExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)`
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}
	return exists, nil
}

func (r *PGRepository) CreateUser(ctx context.Context, user *User) (int64, error) {
	query := `
		INSERT INTO users (email, password_hash, name, is_email_verified, google_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	var userID int
	err := r.db.QueryRowContext(ctx,
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
	return int64(userID), nil
}

func (r *PGRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, password_hash, name, is_email_verified, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	user := &User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
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

func (r *PGRepository) MarkEmailAsVerified(ctx context.Context, userID int64) error {
	query := `
		UPDATE users
		SET is_email_verified = TRUE, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to mark email as verified: %w", err)
	}
	return nil
}

func (r *PGRepository) GetByID(ctx context.Context, id int64) (*User, error) {
	const q = `
        SELECT id, email, password_hash, name, is_email_verified, google_id
        FROM users
        WHERE id = $1
    `
	u := &User{}
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.Name,
		&u.IsEmailVerified,
		&u.GoogleID,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository.GetByID: %w", err)
	}
	return u, nil
}

func (r *PGRepository) GetByGoogleID(ctx context.Context, googleID string) (*User, error) {
	const q = `
        SELECT id, email, password_hash, name, is_email_verified, google_id
        FROM users
        WHERE google_id = $1
    `
	u := &User{}
	err := r.db.QueryRowContext(ctx, q, googleID).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.Name,
		&u.IsEmailVerified,
		&u.GoogleID,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository.GetByGoogleID: %w", err)
	}
	return u, nil
}

func (r *PGRepository) CreateWithGoogle(ctx context.Context, email, googleID string) (int64, error) {
	const q = `
        INSERT INTO users (email, name, is_email_verified, google_id, created_at, updated_at)
        VALUES ($1, '', TRUE, $2, NOW(), NOW())
        RETURNING id
    `
	var newID int64
	if err := r.db.QueryRowContext(ctx, q, email, googleID).Scan(&newID); err != nil {
		return 0, fmt.Errorf("repository.CreateWithGoogle: %w", err)
	}
	return newID, nil
}

func (r *PGRepository) LinkGoogleID(ctx context.Context, userID int64, googleID string) error {
	const q = `
        UPDATE users
        SET google_id = $1, updated_at = NOW()
        WHERE id = $2
    `
	if _, err := r.db.ExecContext(ctx, q, googleID, userID); err != nil {
		return fmt.Errorf("repository.LinkGoogleID: %w", err)
	}
	return nil
}
