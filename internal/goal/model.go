package goal

import (
	"github.com/google/uuid"
	"time"
)

type Goal struct {
	ID            uuid.UUID `json:"id"`
	UserId        int64     `json:"user_id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Status        string    `json:"status"`
	EstimatedTime int64     `json:"estimated_time"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Phase struct {
	ID            uuid.UUID `json:"id"`
	GoalId        uuid.UUID `json:"goalId"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Status        string    `json:"status"`
	EstimatedTime int64     `json:"estimated_time"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Task struct {
	ID            uuid.UUID  `json:"id"`
	GoalId        uuid.UUID  `json:"goalId"`
	PhaseId       *uuid.UUID `json:"phase_id,omitempty"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Status        string     `json:"status"`
	EstimatedTime int64      `json:"estimated_time"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
