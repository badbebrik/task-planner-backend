package create

import (
	"github.com/google/uuid"
	"time"
)

type TaskResponse struct {
	ID            uuid.UUID  `json:"id"`
	GoalID        uuid.UUID  `json:"goal_id"`
	PhaseID       *uuid.UUID `json:"phase_id,omitempty"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Status        string     `json:"status"`
	EstimatedTime int        `json:"estimated_time"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
