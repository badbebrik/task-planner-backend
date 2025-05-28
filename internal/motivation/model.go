package motivation

import (
	"github.com/google/uuid"
	"time"
)

type Motivation struct {
	ID        uuid.UUID
	UserID    int64
	Date      time.Time
	Text      string
	CreatedAt time.Time
}
