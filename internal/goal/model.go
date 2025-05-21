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
	ID            uuid.UUID  `json:"id"`
	GoalId        uuid.UUID  `json:"goalId"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Status        string     `json:"status"` // "not_started", "in_progress", "completed"
	EstimatedTime int        `json:"estimated_time"`
	Progress      int        `json:"progress"`
	Order         int        `json:"order"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
}

type Task struct {
	ID            uuid.UUID  `json:"id"`
	GoalId        uuid.UUID  `json:"goalId"`
	PhaseId       *uuid.UUID `json:"phase_id,omitempty"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Status        string     `json:"status"` // "todo", "in_progress", "completed"
	EstimatedTime int        `json:"estimated_time"`
	TimeSpent     int        `json:"time_spent"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (p *Phase) CalculateProgress(tasks []Task) int {
	spent := 0
	for _, t := range tasks {
		if t.PhaseId != nil && *t.PhaseId == p.ID {
			spent += t.TimeSpent
		}
	}
	if p.EstimatedTime == 0 {
		return 0
	}
	perc := spent * 100 / p.EstimatedTime
	if perc > 100 {
		perc = 100
	}
	return perc
}

func (g *Goal) CalculateProgress(tasks []Task) int {
	spent := 0
	for _, t := range tasks {
		if t.GoalId == g.ID {
			spent += t.TimeSpent
		}
	}
	if g.EstimatedTime == 0 {
		return 0
	}
	perc := spent * 100 / g.EstimatedTime
	if perc > 100 {
		perc = 100
	}
	return perc
}

func (t *Task) CalculateProgress() int {
	if t.EstimatedTime == 0 {
		return 0
	}
	p := t.TimeSpent * 100 / t.EstimatedTime
	if p > 100 {
		p = 100
	}
	return p
}

func (p *Phase) markStarted() {
	if p.StartedAt == nil {
		now := time.Now()
		p.StartedAt = &now
	}
}

func (p *Phase) markCompleted() {
	if p.CompletedAt == nil {
		now := time.Now()
		p.CompletedAt = &now
	}
}
