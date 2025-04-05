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
				VALUES ($1, $2, $3, $4)
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

}

func (r repositoryImpl) ListTimeSlotsByAvailabilityIDs(ctx context.Context, avIDs []uuid.UUID) ([]TimeSlot, error) {
	//TODO implement me
	panic("implement me")
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
