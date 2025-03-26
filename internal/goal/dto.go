package goal

import (
	"github.com/google/uuid"
	"time"
)

type CreateGoalRequest struct {
	Title       string `json:"title" validate:"required,max=255"`
	Description string `json:"description" validate:"required"`
}

type UpdateGoalRequest struct {
	Title       string `json:"title" validate:"omitempty,max=255"`
	Description string `json:"description" validate:"omitempty"`
	Status      string `json:"status" validate:"omitempty,oneof=in-progress completed cancelled"`
}

type CreatePhaseRequest struct {
	GoalID        uuid.UUID `json:"goal_id" validate:"required"`
	Title         string    `json:"title" validate:"required,max=255"`
	Description   string    `json:"description" validate:"required"`
	EstimatedTime int64     `json:"estimated_time" validate:"required,min=0"`
}

type CreateTaskRequest struct {
	GoalID        uuid.UUID  `json:"goal_id" validate:"required"`
	PhaseID       *uuid.UUID `json:"phase_id,omitempty"`
	Title         string     `json:"title" validate:"required,max=255"`
	Description   string     `json:"description" validate:"required"`
	EstimatedTime int64      `json:"estimated_time" validate:"required,min=0"`
}

type UpdateTaskRequest struct {
	Title         string `json:"title" validate:"omitempty,max=255"`
	Description   string `json:"description" validate:"omitempty"`
	Status        string `json:"status" validate:"omitempty,oneof=in-progress completed cancelled"`
	EstimatedTime int64  `json:"estimated_time" validate:"omitempty,min=0"`
}

type GoalResponse struct {
	ID            uuid.UUID       `json:"id"`
	UserID        int64           `json:"user_id"`
	Title         string          `json:"title"`
	Description   string          `json:"description"`
	Status        string          `json:"status"`
	EstimatedTime int64           `json:"estimated_time"`
	Progress      float64         `json:"progress"`
	Phases        []PhaseResponse `json:"phases,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type PhaseResponse struct {
	ID            uuid.UUID      `json:"id"`
	GoalID        uuid.UUID      `json:"goal_id"`
	Title         string         `json:"title"`
	Description   string         `json:"description"`
	Status        string         `json:"status"`
	EstimatedTime int64          `json:"estimated_time"`
	Progress      float64        `json:"progress"`
	Tasks         []TaskResponse `json:"tasks,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type TaskResponse struct {
	ID            uuid.UUID  `json:"id"`
	GoalID        uuid.UUID  `json:"goal_id"`
	PhaseID       *uuid.UUID `json:"phase_id,omitempty"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Status        string     `json:"status"`
	EstimatedTime int64      `json:"estimated_time"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type ListGoalsRequest struct {
	Page     int    `query:"page" validate:"required,min=1"`
	PageSize int    `query:"page_size" validate:"required,min=1,max=100"`
	Status   string `query:"status" validate:"omitempty,oneof=in-progress completed cancelled"`
	SortBy   string `query:"sort_by" validate:"omitempty,oneof=created_at updated_at title status"`
	Order    string `query:"order" validate:"omitempty,oneof=asc desc"`
}

type ListGoalsResponse struct {
	Goals      []GoalResponse `json:"goals"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

type PreviewPhase struct {
	Title         string        `json:"title"`
	Description   string        `json:"description"`
	EstimatedTime int64         `json:"estimated_time"`
	Tasks         []PreviewTask `json:"tasks"`
}

type PreviewTask struct {
	Title         string `json:"title"`
	Description   string `json:"description"`
	EstimatedTime int64  `json:"estimated_time"`
}

type PreviewGoalResponse struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Phases      []PreviewPhase `json:"phases"`
}
