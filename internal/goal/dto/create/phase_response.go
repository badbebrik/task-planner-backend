package create

import (
	"github.com/google/uuid"
	"time"
)

type PhaseResponse struct {
	ID          uuid.UUID      `json:"id"`
	GoalID      uuid.UUID      `json:"goal_id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Status      string         `json:"status"`
	Progress    int            `json:"progress"`
	Order       int            `json:"order"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	Tasks       []TaskResponse `json:"tasks,omitempty"`
}
