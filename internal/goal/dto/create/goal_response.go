package create

import (
	"github.com/google/uuid"
	"time"
)

type GoalResponse struct {
	ID            uuid.UUID       `json:"id"`
	UserID        int64           `json:"user_id"`
	Title         string          `json:"title"`
	Description   string          `json:"description"`
	Status        string          `json:"status"`
	HoursPerWeek  int             `json:"hours_per_week"`
	EstimatedTime int64           `json:"estimated_time"`
	Progress      int             `json:"progress"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	Phases        []PhaseResponse `json:"phases,omitempty"`
}
