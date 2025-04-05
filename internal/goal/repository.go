package goal

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"time"
)

type RepositoryAggregator interface {
	GoalRepository
	PhaseRepository
	TaskRepository
}

type GoalRepository interface {
	CreateGoal(ctx context.Context, g *Goal) error
	GetGoalByID(ctx context.Context, id uuid.UUID) (*Goal, error)
	UpdateGoal(ctx context.Context, g *Goal) error
	ListGoals(ctx context.Context, userID int64, limit, offset int, status string) ([]Goal, int, error)
}
type PhaseRepository interface {
	CreatePhase(ctx context.Context, p *Phase) error
	ListPhasesByGoalID(ctx context.Context, goalID uuid.UUID) ([]Phase, error)
}
type TaskRepository interface {
	CreateTask(ctx context.Context, t *Task) error
	ListTasksByGoalID(ctx context.Context, goalID uuid.UUID) ([]Task, error)
}

type repositoryImpl struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *repositoryImpl {
	return &repositoryImpl{db: db}
}

func (r *repositoryImpl) CreateGoal(ctx context.Context, g *Goal) error {
	query := `
	INSERT INTO goals (id, user_id, title, description, status, estimated_time, hours_per_week, 
    	progress, created_at, updated_at
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.db.ExecContext(ctx, query,
		g.ID,
		g.UserId,
		g.Title,
		g.Description,
		g.Status,
		g.EstimatedTime,
		g.HoursPerWeek,
		g.Progress,
		g.CreatedAt,
		g.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert goal: %w", err)
	}
	return nil
}

func (r *repositoryImpl) GetGoalByID(ctx context.Context, id uuid.UUID) (*Goal, error) {
	query := ` SELECT id, user_id, title, description, status, estimated_time, hours_per_week,
        progress, created_at, updated_at FROM goals
		WHERE id = $1
`
	var g Goal
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(
			&g.ID,
			&g.UserId,
			&g.Title,
			&g.Description,
			&g.Status,
			&g.EstimatedTime,
			&g.HoursPerWeek,
			&g.Progress,
			&g.CreatedAt,
			&g.UpdatedAt,
		)
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
	query := `UPDATE goals
			SET title = $2, description = $3, status = $4, estimated_time = $5, 
				hours_per_week = $6, progress = $7, updated_at = $8
			WHERE id = $
`
	_, err := r.db.ExecContext(ctx, query,
		g.ID,
		g.Title,
		g.Description,
		g.Status,
		g.EstimatedTime,
		g.HoursPerWeek,
		g.Progress,
		g.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update goal: %w", err)
	}
	return nil
}

func (r *repositoryImpl) ListGoals(ctx context.Context, userID int64, limit, offset int, status string) ([]Goal, int, error) {
	countQuery := `SELECT COUNT(*) FROM goals WHERE user_id = $1`
	args := []interface{}{userID}

	if status != "" {
		countQuery += ` AND status = $2`
	}

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count goals: %w", err)
	}

	selectQuery := `SELECT id, user_id, title, description, status, estimated_time, hours_per_week, progress, 
    				created_at, updated_at FROM goals
					WHERE user_id = $1
`
	if status != "" {
		selectQuery += ` AND status = $2`
	}
	selectQuery += ` ORDER BY updated_at DESC LIMIT $3 OFFSET $4`

	var rows *sql.Rows

	if status == "" {
		rows, err = r.db.QueryContext(ctx, selectQuery, userID, limit, offset)
	} else {
		rows, err = r.db.QueryContext(ctx, selectQuery, userID, status, limit, offset)
	}

	if err != nil {
		return nil, 0, fmt.Errorf("failed to query goals: %w", err)
	}
	defer rows.Close()

	var goals []Goal
	for rows.Next() {
		var g Goal
		if err := rows.Scan(
			&g.ID,
			&g.UserId,
			&g.Title,
			&g.Description,
			&g.Status,
			&g.EstimatedTime,
			&g.HoursPerWeek,
			&g.Progress,
			&g.CreatedAt,
			&g.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan goal: %w", err)
		}
		goals = append(goals, g)
	}
	return goals, total, nil
}

func (r *repositoryImpl) CreatePhase(ctx context.Context, p *Phase) error {
	query := `
INSERT INTO phases (id, goal_id, title, description, status, progress, "order",created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`
	_, err := r.db.ExecContext(ctx, query,
		p.ID,
		p.GoalId,
		p.Title,
		p.Description,
		p.Status,
		p.Progress,
		p.Order,
		p.CreatedAt,
		p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert phase: %w", err)
	}
	return nil
}

func (r *repositoryImpl) ListPhasesByGoalID(ctx context.Context, goalID uuid.UUID) ([]Phase, error) {
	query := `SELECT id, goal_id, title, description, status, progress, "order",
    			created_at, updated_at FROM phases WHERE goal_id = $1
				ORDER BY "order" ASC
`
	rows, err := r.db.QueryContext(ctx, query, goalID)
	if err != nil {
		return nil, fmt.Errorf("failed to list phases: %w", err)
	}
	defer rows.Close()

	var phases []Phase
	for rows.Next() {
		var p Phase
		if err := rows.Scan(
			&p.ID,
			&p.GoalId,
			&p.Title,
			&p.Description,
			&p.Status,
			&p.Progress,
			&p.Order,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan phase: %w", err)
		}
		phases = append(phases, p)
	}
	return phases, nil
}

func (r *repositoryImpl) CreateTask(ctx context.Context, t *Task) error {
	query := `
INSERT INTO tasks (id, goal_id, phase_id, title, description, status, estimated_time, completed_at, 
    			created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
`
	_, err := r.db.ExecContext(ctx, query,
		t.ID,
		t.GoalId,
		t.PhaseId,
		t.Title,
		t.Description,
		t.Status,
		t.EstimatedTime,
		t.CompletedAt,
		t.CreatedAt,
		t.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert task: %w", err)
	}
	return nil
}

func (r *repositoryImpl) ListTasksByGoalID(ctx context.Context, goalID uuid.UUID) ([]Task, error) {
	query := `SELECT id, goal_id, phase_id, title, description, status, estimated_time, 
    		completed_at, created_at, updated_at FROM tasks WHERE goal_id = $1
			ORDER BY created_at ASC
`
	rows, err := r.db.QueryContext(ctx, query, goalID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(
			&t.ID,
			&t.GoalId,
			&t.PhaseId,
			&t.Title,
			&t.Description,
			&t.Status,
			&t.EstimatedTime,
			&t.CompletedAt,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}
