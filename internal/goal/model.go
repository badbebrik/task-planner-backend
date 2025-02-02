package goal

import "time"

type Goal struct {
	ID            int64     `json:"id"`
	UserId        int64     `json:"user_id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Status        string    `json:"status"`
	EstimatedTime int64     `json:"estimated_time"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Phase struct {
	ID            int64     `json:"id"`
	GoalId        int64     `json:"goalId"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Status        string    `json:"status"`
	EstimatedTime int64     `json:"estimated_time"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Task struct {
	ID            int64     `json:"id"`
	GoalId        int64     `json:"goalId"`
	PhaseId       *int64    `json:"phase_id,omitempty"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Status        string    `json:"status"`
	EstimatedTime int64     `json:"estimated_time"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
