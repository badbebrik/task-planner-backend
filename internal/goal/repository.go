package goal

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type GoalRepository interface {
	CreateGoal(ctx context.Context, g *Goal) error
	GetGoalByID(ctx context.Context, id int64) (*Goal, error)
	UpdateGoal(ctx context.Context, g *Goal) error
}
type PhaseRepository interface {
	CreatePhase(ctx context.Context, p *Phase) error
}
type TaskRepository interface {
	CreateTask(ctx context.Context, t *Task) error
}

type repositoryImpl struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *repositoryImpl {
	return &repositoryImpl{db: db}
}

func (r *repositoryImpl) CreateGoal(ctx context.Context, g *Goal) error {
	query := `
        INSERT INTO goals (id, user_id, title, description, status, estimated_time, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `
	_, err := r.db.ExecContext(ctx, query,
		g.ID, g.UserId, g.Title, g.Description, g.Status, g.EstimatedTime, time.Now(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to insert goal: %w", err)
	}
	return nil
}

func (r *repositoryImpl) GetGoalByID(ctx context.Context, goalID string) (*Goal, error) {
	query := `
        SELECT id, user_id, title, description, status, estimated_time, created_at, updated_at
        FROM goals
        WHERE id = $1
    `
	var g Goal
	err := r.db.QueryRowContext(ctx, query, goalID).
		Scan(&g.ID, &g.UserId, &g.Title, &g.Description, &g.Status, &g.EstimatedTime, &g.CreatedAt, &g.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get goal by id: %w", err)
	}
	return &g, nil
}

func (r *repositoryImpl) UpdateGoal(ctx context.Context, g *Goal) error {
	g.UpdatedAt = time.Now()
	query := `
        UPDATE goals
        SET title = $2, description = $3, status = $4, estimated_time = $5, updated_at = $6
        WHERE id = $1
    `
	_, err := r.db.ExecContext(ctx, query,
		g.ID, g.Title, g.Description, g.Status, g.EstimatedTime, g.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update goal: %w", err)
	}
	return nil
}

func (r *repositoryImpl) CreatePhase(ctx context.Context, p *Phase) error {
	query := `
        INSERT INTO phases (id, goal_id, title, description, status, estimated_time, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `
	_, err := r.db.ExecContext(ctx, query,
		p.ID, p.GoalId, p.Title, p.Description, p.Status, p.EstimatedTime, time.Now(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to insert phase: %w", err)
	}
	return nil
}

func (r *repositoryImpl) CreateTask(ctx context.Context, t *Task) error {
	query := `
        INSERT INTO tasks (id, goal_id, phase_id, title, description, status, estimated_time, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    `
	_, err := r.db.ExecContext(ctx, query,
		t.ID, t.GoalId, t.PhaseId, t.Title, t.Description, t.Status, t.EstimatedTime, time.Now(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to insert task: %w", err)
	}
	return nil
}
