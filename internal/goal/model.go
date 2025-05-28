package goal

import (
	"github.com/google/uuid"
	"log"
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
			log.Printf("[PHASE PROGRESS] Task with id: %s, Time Spent: %d, PhaseID: %s", t.ID, t.TimeSpent, t.PhaseId)
			spent += t.TimeSpent
		}
	}
	log.Printf("Phase with id: %d Estimated time of phase: %d", p.ID, p.EstimatedTime)
	if p.EstimatedTime == 0 {
		return 0
	}
	log.Printf("[PHASE PROGRESS] Spent: %d, EstimatedTime * 60: %d", spent, p.EstimatedTime*60)
	perc := spent * 100 / (p.EstimatedTime * 60)
	if perc > 100 {
		perc = 100
	}
	log.Printf("[PHASE PROGRESS] Perc of phase %s : %d", p.ID, perc)
	return perc
}

func (g *Goal) CalculateProgress(tasks []Task) int {
	spent := 0
	for _, t := range tasks {
		if t.GoalId == g.ID {
			log.Printf("[PHASE PROGRESS] Task with id: %s, Time Spent: %d, PhaseID: %s", t.ID, t.TimeSpent, t.PhaseId)
			spent += t.TimeSpent
			log.Printf("[GOAL PROGRESS] Spent Total: %d", spent)
		}
	}
	if g.EstimatedTime == 0 {
		return 0
	}
	log.Printf("[GOAL PROGRESS] Spent: %d, EstimatedTime * 60: %d", spent, g.EstimatedTime*60)
	perc := spent * 100 / (g.EstimatedTime * 60)
	if perc > 100 {
		perc = 100
	}
	log.Printf("[GOAL PROGRESS] Perc: %d", perc)
	return perc
}

func (t *Task) CalculateProgress() int {
	if t.EstimatedTime == 0 {
		return 0
	}
	log.Printf("[TASK PROGRESS] Spent: %d, EstimatedTime: %d", t.TimeSpent, t.EstimatedTime*60)
	p := t.TimeSpent * 100 / (t.EstimatedTime * 60)
	if p > 100 {
		p = 100
	}
	log.Printf("[TASK PROGRESS] Perc: %d", p)
	return p
}

func (p *Phase) MarkStarted() {
	if p.StartedAt == nil {
		now := time.Now()
		p.StartedAt = &now
	}
}

func (p *Phase) MarkCompleted() {
	if p.CompletedAt == nil {
		now := time.Now()
		p.CompletedAt = &now
	}
}
