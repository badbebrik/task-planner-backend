package motivation

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Repository interface {
	Create(ctx context.Context, m *Motivation) error
	GetByUserAndDate(ctx context.Context, userID int64, date time.Time) (*Motivation, error)
}

type repositoryImpl struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repositoryImpl{db: db}
}

func (r *repositoryImpl) Create(ctx context.Context, m *Motivation) error {
	query := `
INSERT INTO motivation (id, user_id, date, text, created_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id, date) DO NOTHING
`
	_, err := r.db.ExecContext(ctx, query,
		m.ID, m.UserID, m.Date.Format("2006-01-02"), m.Text, m.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert motivation: %w", err)
	}
	return nil
}

func (r *repositoryImpl) GetByUserAndDate(ctx context.Context, userID int64, date time.Time) (*Motivation, error) {
	query := `
SELECT id, user_id, date, text, created_at
FROM motivation
WHERE user_id = $1 AND date = $2
`
	var m Motivation
	row := r.db.QueryRowContext(ctx, query, userID, date.Format("2006-01-02"))
	if err := row.Scan(&m.ID, &m.UserID, &m.Date, &m.Text, &m.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query motivation: %w", err)
	}
	return &m, nil
}
