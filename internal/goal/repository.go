package goal

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"strconv"
	"strings"
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

	GetPhaseByID(ctx context.Context, id uuid.UUID) (*Phase, error)
	UpdatePhase(ctx context.Context, p *Phase) error
	GetTaskByID(ctx context.Context, id uuid.UUID) (*Task, error)
	UpdateTask(ctx context.Context, t *Task) error
	UpdateTaskTimeSpent(ctx context.Context, id uuid.UUID, spent int) error
	DeleteGoal(ctx context.Context, id uuid.UUID) error
}
type PhaseRepository interface {
	CreatePhase(ctx context.Context, p *Phase) error
	ListPhasesByGoalID(ctx context.Context, goalID uuid.UUID) ([]Phase, error)
}
type TaskRepository interface {
	CreateTask(ctx context.Context, t *Task) error
	ListTasksByGoalID(ctx context.Context, goalID uuid.UUID) ([]Task, error)
	GetTasksByIDs(ctx context.Context, ids []uuid.UUID) ([]Task, error)
	GetGoalsByIDs(ctx context.Context, ids []uuid.UUID) ([]Goal, error)
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

func (r *repositoryImpl) ListGoals(
	ctx context.Context,
	userID int64,
	limit, offset int,
	status string,
) ([]Goal, int, error) {
	args := make([]interface{}, 0, 4)
	idx := 1

	where := "WHERE user_id = $" + strconv.Itoa(idx)
	args = append(args, userID)
	idx++

	if status != "" {
		where += " AND status = $" + strconv.Itoa(idx)
		args = append(args, status)
		idx++
	}

	countQuery := "SELECT COUNT(*) FROM goals " + where
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count goals: %w", err)
	}

	selectQuery := "" +
		"SELECT id, user_id, title, description, status, estimated_time, hours_per_week, progress, " +
		"       created_at, updated_at " +
		"FROM goals " + where +
		" ORDER BY updated_at DESC " +
		" LIMIT $" + strconv.Itoa(idx) +
		" OFFSET $" + strconv.Itoa(idx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, selectQuery, args...)
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
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("row iteration error: %w", err)
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

func (r *repositoryImpl) GetTasksByIDs(ctx context.Context, ids []uuid.UUID) ([]Task, error) {
	if len(ids) == 0 {
		return []Task{}, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	query := fmt.Sprintf(`
        SELECT id, goal_id, phase_id, title, description, status, estimated_time, created_at, updated_at
        FROM tasks
        WHERE id IN (%s)
    `, strings.Join(placeholders, ", "))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query tasks by ids: %w", err)
	}
	defer rows.Close()

	var result []Task
	for rows.Next() {
		var t Task
		err := rows.Scan(
			&t.ID, &t.GoalId, &t.PhaseId, &t.Title, &t.Description,
			&t.Status, &t.EstimatedTime, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		result = append(result, t)
	}

	return result, nil
}

func (r *repositoryImpl) GetGoalsByIDs(ctx context.Context, ids []uuid.UUID) ([]Goal, error) {
	if len(ids) == 0 {
		return []Goal{}, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	query := fmt.Sprintf(`
        SELECT id, user_id, title, description, status, estimated_time, created_at, updated_at
        FROM goals
        WHERE id IN (%s)
    `, strings.Join(placeholders, ", "))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query goals by ids: %w", err)
	}
	defer rows.Close()

	var result []Goal
	for rows.Next() {
		var g Goal
		err := rows.Scan(
			&g.ID, &g.UserId, &g.Title, &g.Description, &g.Status,
			&g.EstimatedTime, &g.CreatedAt, &g.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan goal: %w", err)
		}
		result = append(result, g)
	}

	return result, nil
}

func (r *repositoryImpl) GetTaskByID(ctx context.Context, id uuid.UUID) (*Task, error) {
	q := `SELECT id, goal_id, phase_id, title, description, status,
	             estimated_time, time_spent, completed_at, created_at, updated_at
	      FROM tasks WHERE id = $1`
	var t Task
	var phaseID *uuid.UUID
	if err := r.db.QueryRowContext(ctx, q, id).Scan(
		&t.ID, &t.GoalId, &phaseID, &t.Title, &t.Description,
		&t.Status, &t.EstimatedTime, &t.TimeSpent,
		&t.CompletedAt, &t.CreatedAt, &t.UpdatedAt); err != nil {
		return nil, err
	}
	t.PhaseId = phaseID
	return &t, nil
}

func (r *repositoryImpl) UpdateTask(ctx context.Context, t *Task) error {
	t.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx, `
	    UPDATE tasks
	    SET status = $2, time_spent = $3, completed_at = $4, updated_at = $5
	    WHERE id = $1`,
		t.ID, t.Status, t.TimeSpent, t.CompletedAt, t.UpdatedAt)
	return err
}

func (r *repositoryImpl) UpdateTaskTimeSpent(ctx context.Context, id uuid.UUID, spent int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE tasks SET time_spent = $2, updated_at = NOW() WHERE id = $1`, id, spent)
	return err
}

func (r *repositoryImpl) GetPhaseByID(ctx context.Context, id uuid.UUID) (*Phase, error) {
	q := `SELECT id, goal_id, title, description, status,
	             estimated_time, progress, "order", created_at, updated_at
	      FROM phases WHERE id = $1`
	var p Phase
	if err := r.db.QueryRowContext(ctx, q, id).Scan(
		&p.ID, &p.GoalId, &p.Title, &p.Description, &p.Status,
		&p.EstimatedTime, &p.Progress, &p.Order, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *repositoryImpl) UpdatePhase(ctx context.Context, p *Phase) error {
	p.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx, `
	    UPDATE phases
	    SET status = $2, progress = $3, updated_at = $4, started_at = $5, completed_at = $6
	    WHERE id = $1`,
		p.ID, p.Status, p.Progress, p.UpdatedAt, p.StartedAt, p.CompletedAt)
	return err
}

func (r *repositoryImpl) DeleteGoal(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM goals WHERE id = $1`, id)
	return err
}
