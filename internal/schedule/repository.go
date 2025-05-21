package schedule

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"log"
	"time"
)

type Repository interface {
	DeleteAvailabilityByGoal(ctx context.Context, goalID uuid.UUID) error
	CreateAvailability(ctx context.Context, av *Availability) error
	ListAvailabilityByGoal(ctx context.Context, goalID uuid.UUID) ([]Availability, error)

	CreateTimeSlot(ctx context.Context, slot *TimeSlot) error
	ListTimeSlotsByAvailabilityIDs(ctx context.Context, avIDs []uuid.UUID) ([]TimeSlot, error)

	CreateScheduledTask(ctx context.Context, st *ScheduledTask) error
	DeleteScheduledTasksByGoal(ctx context.Context, goalID uuid.UUID) error
	ListScheduledTasksForDate(ctx context.Context, date time.Time) ([]ScheduledTask, error)
	ListScheduledTasksInRange(ctx context.Context, startDate, endDate time.Time) ([]ScheduledTask, error)
	ListUpcomingTasks(ctx context.Context, limit int) ([]ScheduledTask, error)

	ListScheduledTasksForGoalInRange(ctx context.Context, goalID uuid.UUID, startDate, endDate time.Time) ([]ScheduledTask, error)

	// todo: дополнить для статы или выкинуть нафиг
	CountTasksByDay(ctx context.Context, startDate, endDate time.Time) (map[time.Time]struct{ Completed, Total int }, error)

	UpdateScheduledTaskStatus(ctx context.Context, id uuid.UUID, newStatus string) error
	GetScheduledTaskByID(ctx context.Context, id uuid.UUID) (*ScheduledTask, error)
	SumDoneIntervalsForTask(ctx context.Context, taskID uuid.UUID) (int, error)
}

type repositoryImpl struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repositoryImpl{db: db}
}

func (r *repositoryImpl) DeleteAvailabilityByGoal(ctx context.Context, goalID uuid.UUID) error {
	query := `DELETE FROM availability WHERE goal_id = $1`
	_, err := r.db.ExecContext(ctx, query, goalID)
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

func (r repositoryImpl) ListAvailabilityByGoal(ctx context.Context, goalID uuid.UUID) ([]Availability, error) {
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
	query := `INSERT INTO time_slot (id, availability_id, start_time, end_time, created_at, updated_at)
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

func (r repositoryImpl) ListTimeSlotsByAvailabilityIDs(
	ctx context.Context, avIDs []uuid.UUID,
) ([]TimeSlot, error) {

	const query = `
      SELECT id, availability_id, start_time, end_time,
             created_at, updated_at
      FROM time_slot
      WHERE availability_id = ANY($1)
      ORDER BY start_time
    `

	rows, err := r.db.QueryContext(ctx, query, pq.Array(avIDs))
	if err != nil {
		return nil, fmt.Errorf("failed to list time_slots: %w", err)
	}
	defer rows.Close()

	var result []TimeSlot
	for rows.Next() {
		var ts TimeSlot
		if err := rows.Scan(
			&ts.ID,
			&ts.AvailabilityID,
			&ts.StartTime,
			&ts.EndTime,
			&ts.CreatedAt,
			&ts.UpdatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, ts)
	}
	return result, nil
}

func (r repositoryImpl) CreateScheduledTask(ctx context.Context, st *ScheduledTask) error {
	query := `INSERT INTO scheduled_task (id, task_id, time_slot_id, scheduled_date, start_time, end_time, status, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`
	_, err := r.db.ExecContext(ctx, query,
		st.ID,
		st.TaskID,
		st.TimeSlotID,
		st.ScheduledDate,
		st.StartTime,
		st.EndTime,
		st.Status,
		st.CreatedAt,
		st.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create schedued_task: %w", err)
	}

	return nil
}

func (r repositoryImpl) DeleteScheduledTasksByGoal(ctx context.Context, goalID uuid.UUID) error {
	query := `DELETE FROM scheduled_task USING tasks WHERE scheduled_task.task_id = tasks.id AND task.goal_id = $1

`
	_, err := r.db.ExecContext(ctx, query, goalID)
	if err != nil {
		return fmt.Errorf("failed to delete scheduled tasks by goal: %w", err)
	}
	return nil
}

func (r repositoryImpl) ListScheduledTasksForDate(ctx context.Context, date time.Time) ([]ScheduledTask, error) {
	query := `SELECT st.id, st.task_id, st.time_slot_id, st.scheduled_date, st.start_time, st.end_time, st.status, st.created_at, st.updated_at
		FROM scheduled_task st WHERE st.scheduled_date = $1 ORDER BY st.start_time
`
	rows, err := r.db.QueryContext(ctx, query, date.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("failed to list scheduled tasks for date: %w", err)
	}
	defer rows.Close()

	var result []ScheduledTask
	for rows.Next() {
		var st ScheduledTask
		var dateStr, stStr, etStr string
		if err := rows.Scan(
			&st.ID,
			&st.TaskID,
			&st.TimeSlotID,
			&dateStr,
			&stStr,
			&etStr,
			&st.Status,
			&st.CreatedAt,
			&st.UpdatedAt,
		); err != nil {
			return nil, err
		}
		sd, _ := time.Parse("2006-01-02", dateStr)
		stt, _ := time.Parse("15:04:05", stStr)
		ett, _ := time.Parse("15:04:05", etStr)

		st.ScheduledDate = sd
		st.StartTime = combineDateTime(sd, stt)
		st.EndTime = combineDateTime(sd, ett)
		result = append(result, st)
	}
	return result, nil
}

func combineDateTime(date, tm time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), tm.Hour(), tm.Minute(), tm.Second(), 0, time.UTC)
}

func (r repositoryImpl) ListScheduledTasksInRange(ctx context.Context, startDate, endDate time.Time) ([]ScheduledTask, error) {
	query := `
SELECT 
    st.id, st.task_id, st.time_slot_id,
    st.scheduled_date, st.start_time, st.end_time,
    st.status, st.created_at, st.updated_at
FROM scheduled_task st
WHERE st.scheduled_date >= $1
  AND st.scheduled_date <= $2
ORDER BY st.scheduled_date, st.start_time
`
	rows, err := r.db.QueryContext(ctx, query,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list scheduled tasks in range: %w", err)
	}
	defer rows.Close()

	var result []ScheduledTask
	for rows.Next() {
		var st ScheduledTask
		var dateOnly, timeStart, timeEnd time.Time
		if err := rows.Scan(
			&st.ID,
			&st.TaskID,
			&st.TimeSlotID,
			&dateOnly,
			&timeStart,
			&timeEnd,
			&st.Status,
			&st.CreatedAt,
			&st.UpdatedAt,
		); err != nil {
			return nil, err
		}
		log.Printf("[Repo][ListScheduledTasksInRange] raw dateStr=%q, startStr=%q, endStr=%q", dateOnly, startDate, endDate)
		//sd, _ := time.Parse("2006-01-02", dateStr)
		//stt, _ := time.Parse("15:04:05", stStr)
		//ett, _ := time.Parse("15:04:05", etStr)

		st.ScheduledDate = dateOnly
		st.StartTime = combineDateTime(dateOnly, timeStart)
		st.EndTime = combineDateTime(dateOnly, timeEnd)

		result = append(result, st)
	}
	return result, nil
}

func (r repositoryImpl) ListUpcomingTasks(ctx context.Context, limit int) ([]ScheduledTask, error) {
	query := fmt.Sprintf(`
SELECT 
    st.id, st.task_id, st.time_slot_id,
    st.scheduled_date, st.start_time, st.end_time,
    st.status, st.created_at, st.updated_at
FROM scheduled_task st
WHERE st.scheduled_date >= $1
ORDER BY st.scheduled_date, st.start_time
LIMIT %d
`, limit)

	rows, err := r.db.QueryContext(ctx, query, time.Now().Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("failed to list upcoming tasks: %w", err)
	}
	defer rows.Close()

	var result []ScheduledTask
	for rows.Next() {
		var st ScheduledTask
		var dateStr, stStr, etStr string
		if err := rows.Scan(
			&st.ID,
			&st.TaskID,
			&st.TimeSlotID,
			&dateStr,
			&stStr,
			&etStr,
			&st.Status,
			&st.CreatedAt,
			&st.UpdatedAt,
		); err != nil {
			return nil, err
		}
		sd, _ := time.Parse("2006-01-02", dateStr)
		stt, _ := time.Parse("15:04:05", stStr)
		ett, _ := time.Parse("15:04:05", etStr)

		st.ScheduledDate = sd
		st.StartTime = combineDateTime(sd, stt)
		st.EndTime = combineDateTime(sd, ett)

		result = append(result, st)
	}
	return result, nil
}

func (r repositoryImpl) CountTasksByDay(ctx context.Context, startDate, endDate time.Time) (map[time.Time]struct{ Completed, Total int }, error) {
	query := `
SELECT 
    scheduled_date,
    SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed,
    COUNT(*) as total
FROM scheduled_task
WHERE scheduled_date >= $1
  AND scheduled_date <= $2
GROUP BY scheduled_date
`
	rows, err := r.db.QueryContext(ctx, query,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to count tasks by day: %w", err)
	}
	defer rows.Close()

	result := make(map[time.Time]struct{ Completed, Total int })
	for rows.Next() {
		var dateStr string
		var completed, total int
		if err := rows.Scan(&dateStr, &completed, &total); err != nil {
			return nil, err
		}
		dt, _ := time.Parse("2006-01-02", dateStr)
		result[dt] = struct{ Completed, Total int }{
			Completed: completed,
			Total:     total,
		}
	}
	return result, nil
}

func (r *repositoryImpl) ListScheduledTasksForGoalInRange(ctx context.Context, goalID uuid.UUID, startDate, endDate time.Time) ([]ScheduledTask, error) {
	query := `
SELECT 
    id, 
    task_id, 
    time_slot_id, 
    scheduled_date, 
    start_time, 
    end_time,
    status,
    created_at
FROM scheduled_task
WHERE scheduled_date >= $1
  AND scheduled_date <= $2
  AND task_id IN (
      SELECT id FROM tasks WHERE goal_id = $3
  )
ORDER BY scheduled_date, start_time
`
	rows, err := r.db.QueryContext(ctx,
		query,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
		goalID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list scheduled tasks: %w", err)
	}
	defer rows.Close()

	var result []ScheduledTask
	for rows.Next() {
		var st ScheduledTask
		var dateStr, startStr, endStr string
		if err := rows.Scan(
			&st.ID,
			&st.TaskID,
			&st.TimeSlotID,
			&dateStr,
			&startStr,
			&endStr,
			&st.Status,
			&st.CreatedAt,
		); err != nil {
			return nil, err
		}

		sd, _ := time.Parse("2006-01-02", dateStr)
		stt, _ := time.Parse("15:04:05", startStr)
		ett, _ := time.Parse("15:04:05", endStr)

		st.ScheduledDate = sd
		st.StartTime = time.Date(sd.Year(), sd.Month(), sd.Day(), stt.Hour(), stt.Minute(), stt.Second(), 0, time.UTC)
		st.EndTime = time.Date(sd.Year(), sd.Month(), sd.Day(), ett.Hour(), ett.Minute(), ett.Second(), 0, time.UTC)

		result = append(result, st)
	}

	return result, nil
}

func (r *repositoryImpl) UpdateScheduledTaskStatus(ctx context.Context, id uuid.UUID, newStatus string) error {
	query := ` UPDATE scheduled_task
			 SET status = $2, updated_at = now()
			 WHERE id = $1;
`
	_, err := r.db.ExecContext(ctx, query, newStatus, id)
	if err != nil {
		return fmt.Errorf("failed to update scheduled tasks status: %w", err)
	}
	return nil
}

func (r repositoryImpl) GetScheduledTaskByID(
	ctx context.Context, id uuid.UUID,
) (*ScheduledTask, error) {
	q := `SELECT id, task_id, time_slot_id,
	      scheduled_date, start_time, end_time, status, created_at, updated_at
	      FROM scheduled_task WHERE id = $1`
	var st ScheduledTask
	var dateStr, stStr, etStr string
	if err := r.db.QueryRowContext(ctx, q, id).Scan(
		&st.ID, &st.TaskID, &st.TimeSlotID,
		&dateStr, &stStr, &etStr, &st.Status, &st.CreatedAt, &st.UpdatedAt); err != nil {
		return nil, err
	}
	sd, _ := time.Parse("2006-01-02", dateStr)
	stt, _ := time.Parse("15:04:05", stStr)
	ett, _ := time.Parse("15:04:05", etStr)
	st.ScheduledDate = sd
	st.StartTime = combineDateTime(sd, stt)
	st.EndTime = combineDateTime(sd, ett)
	return &st, nil
}

func (r repositoryImpl) SumDoneIntervalsForTask(
	ctx context.Context, taskID uuid.UUID,
) (int, error) {
	q := `SELECT COALESCE(
	          SUM( EXTRACT(EPOCH FROM (end_time - start_time))), 0)
	      FROM scheduled_task
	      WHERE task_id = $1 AND status = 'done'`
	var min int
	if err := r.db.QueryRowContext(ctx, q, taskID).Scan(&min); err != nil {
		return 0, err
	}
	return min, nil
}
