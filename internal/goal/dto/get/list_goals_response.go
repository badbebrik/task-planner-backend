package get

import (
	"github.com/google/uuid"
	"time"
)

type ListGoalsResponse struct {
	Goals []ListGoalItem `json:"goals"`
	Meta  struct {
		Total  int `json:"total"`
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
	} `json:"meta"`
}

type ListGoalItem struct {
	ID           uuid.UUID `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Status       string    `json:"status"`
	Progress     int       `json:"progress"`
	HoursPerWeek int       `json:"hours_per_week"`
	UpdatedAt    time.Time `json:"updated_at"`
	NextTask     *struct {
		ID      uuid.UUID  `json:"id"`
		Title   string     `json:"title"`
		DueDate *time.Time `json:"due_date,omitempty"`
	} `json:"next_task,omitempty"`
}
