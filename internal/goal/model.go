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
	Status        string    `json:"status"` // "planning", "active", "completed", "paused"
	EstimatedTime int       `json:"estimated_time"`
	Progress      int       `json:"progress"`
	HoursPerWeek  int       `json:"hoursPerWeek"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Phase struct {
	ID            uuid.UUID `json:"id"`
	GoalId        uuid.UUID `json:"goalId"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Status        string    `json:"status"` // "not_started", "in_progress", "completed"
	EstimatedTime int       `json:"estimated_time"`
	Progress      int       `json:"progress"`
	Order         int       `json:"order"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Task struct {
	ID            uuid.UUID  `json:"id"`
	GoalId        uuid.UUID  `json:"goalId"`
	PhaseId       *uuid.UUID `json:"phase_id,omitempty"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Status        string     `json:"status"` // "todo", "in_progress", "completed"
	EstimatedTime int        `json:"estimated_time"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (p *Phase) CalculateProgress(tasks []Task) int {
	timeSpent := 0
	for _, t := range tasks {
		if t.PhaseId != nil && *t.PhaseId == p.ID && t.Status == "completed" {
			timeSpent += t.EstimatedTime
		}
	}
	return timeSpent / p.EstimatedTime * 100
}

func (g *Goal) CalculateProgress(tasks []Task) int {
	timeSpent := 0
	for _, t := range tasks {
		if t.GoalId == g.ID && t.Status == "completed" {
			timeSpent += t.EstimatedTime
		}
	}

	result := max(timeSpent/g.EstimatedTime*100, 100)

	return result
}
