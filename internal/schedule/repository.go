package schedule

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"time"
)

type Repository interface {
	DeleteAvailabilityByGoal(ctx context.Context, goalID uuid.UUID) error
	CreateAvailability(ctx context.Context, av *Availability) error
	ListAvailabilityByGoall(ctx context.Context, goalID uuid.UUID) ([]Availability, error)

	CreateTimeSlot(ctx context.Context, slot *TimeSlot) error
	ListTimeSlotsByAvailabilityIDs(ctx context.Context, avIDs []uuid.UUID) ([]TimeSlot, error)

	CreateScheduledTask(ctx context.Context, st *ScheduledTask) error
	DeleteScheduledTasksByGoal(ctx context.Context, goalID uuid.UUID) error
	ListScheduledTasksForDate(ctx context.Context, date time.Time) ([]ScheduledTask, error)
	ListScheduledTasksInRange(ctx context.Context, startDate, endDate time.Time) ([]ScheduledTask, error)
	ListUpcomingTasks(ctx context.Context, limit int) ([]ScheduledTask, error)

	// todo: дополнить для статы или выкинуть нафиг
	CountTasksByDay(ctx context.Context, startDate, endDate time.Time) (map[time.Time]struct{ Completed, Total int }, error)
}

type repositoryImpl struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repositoryImpl{db: db}
}

func (r *repositoryImpl) DeleteAvailabilityByGoal(ctx context.Context, goalID uuid.UUID) error {
	query := `DELETE FROM availability WHERE goal_id = $1`
	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete old availability: %w", err)
	}
	return nil
}

func (r repositoryImpl) CreateAvailability(ctx context.Context, av *Availability) error {
	query := `INSERT INTO availability (id, goal_id, day_of_week, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5)
`
	_, err := r.db.ExecContext(ctx, query, av.ID, av.GoalID, av.DayOfWeek, av.CreatedAt, av.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create availability: %w", err)
	}
	return nil
}

func (r repositoryImpl) ListAvailabilityByGoall(ctx context.Context, goalID uuid.UUID) ([]Availability, error) {
	query := `SELECT id, goal_id, day_of_week, created_at, updated_at FROM availability WHERE goal_id = $1 ORDER BY day_of_week ASC`
	rows, err := r.db.QueryContext(ctx, query, goalID)
	if err != nil {
		return nil, fmt.Errorf("failed to list availability: %w", err)
	}
	defer rows.Close()

	var result []Availability
	for rows.Next() {
		var av Availability
		if err := rows.Scan(&av.ID, &av.GoalID, &av.DayOfWeek, &av.CreatedAt, &av.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, av)
	}
	return result, nil
}

func (r repositoryImpl) CreateTimeSlot(ctx context.Context, slot *TimeSlot) error {
	query := `INSERT INTO time_slot (id, availability_id, start_id, end_time, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
`
	_, err := r.db.ExecContext(ctx, query,
		slot.ID,
		slot.AvailabilityID,
		slot.StartTime,
		slot.EndTime,
		slot.CreatedAt,
		slot.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create time_slot: %w", err)
	}

	return nil
}

func (r repositoryImpl) ListTimeSlotsByAvailabilityIDs(ctx context.Context, avIDs []uuid.UUID) ([]TimeSlot, error) {
	if len(avIDs) == 0 {
		return []TimeSlot{}, nil
	}

	query := `SELECT id, availability_id, start_time, end_time, created_at FROM time_slot WHERE availability_id = ANY ($1) ORDER BY start_time
`
	rows, err := r.db.QueryContext(ctx, query, pqArrayUUID(avIDs))
	if err != nil {
		return nil, fmt.Errorf("failed to list time_slots: %w", err)
	}
	defer rows.Close()

	var result []TimeSlot
	for rows.Next() {
		var ts TimeSlot
		var startStr, endStr string
		if err := rows.Scan(
			&ts.ID,
			&ts.AvailabilityID,
			&ts.StartTime,
			&ts.EndTime,
			&startStr,
			&endStr,
		); err != nil {
			return nil, err
		}

		st, _ := time.Parse("15:04:05", startStr)
		et, _ := time.Parse("15:04:05", endStr)
		ts.StartTime = st
		ts.EndTime = et

		result = append(result, ts)
	}

	return result, nil
}

func pqArrayUUID(list []uuid.UUID) []string {
	out := make([]string, 0, len(list))
	for _, id := range list {
		out = append(out, id.String())
	}
	return out
}

func (r repositoryImpl) CreateScheduledTask(ctx context.Context, st *ScheduledTask) error {
	//TODO implement me
	panic("implement me")
}

func (r repositoryImpl) DeleteScheduledTasksByGoal(ctx context.Context, goalID uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (r repositoryImpl) ListScheduledTasksForDate(ctx context.Context, date time.Time) ([]ScheduledTask, error) {
	//TODO implement me
	panic("implement me")
}

func (r repositoryImpl) ListScheduledTasksInRange(ctx context.Context, startDate, endDate time.Time) ([]ScheduledTask, error) {
	//TODO implement me
	panic("implement me")
}

func (r repositoryImpl) ListUpcomingTasks(ctx context.Context, limit int) ([]ScheduledTask, error) {
	//TODO implement me
	panic("implement me")
}

func (r repositoryImpl) CountTasksByDay(ctx context.Context, startDate, endDate time.Time) (map[time.Time]struct{ Completed, Total int }, error) {
	//TODO implement me
	panic("implement me")
}
