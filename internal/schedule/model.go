package schedule

import (
	"github.com/google/uuid"
	"time"
)

type Availability struct {
	ID        uuid.UUID `json:"id"`
	GoalID    uuid.UUID `json:"goal_id"`
	DayOfWeek int       `json:"day_of_week"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TimeSlot struct {
	ID             uuid.UUID `json:"id"`
	AvailabilityID uuid.UUID `json:"availability_id"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type ScheduledTask struct {
	ID            uuid.UUID `json:"id"`
	TaskID        uuid.UUID `json:"task_id"`
	TimeSlotID    uuid.UUID `json:"time_slot_id"`
	ScheduledDate time.Time `json:"scheduled_date"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
