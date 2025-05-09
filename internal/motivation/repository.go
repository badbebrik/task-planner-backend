package motivation

import (
	"context"
	"time"
)

type Repository interface {
	Create(ctx context.Context, m *Motivation) error
	GetByUserAndDate(ctx context.Context, userID int64, date time.Time) (*Motivation, error)
}
