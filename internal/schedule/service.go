package schedule

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"task-planner/internal/goal"
	"task-planner/internal/schedule/dto"
	"time"
)

type Service interface {
	UpdateAvailability(ctx context.Context, goalID uuid.UUID, req dto.UpdateAvailabilityRequest) (*dto.UpdateAvailabilityResponse, error)
	ListAvailability(ctx context.Context, goalID uuid.UUID) (*dto.UpdateAvailabilityRequest, error)

	AutoScheduleForGoal(ctx context.Context, goalID uuid.UUID) (int, error)

	GetScheduleForDay(ctx context.Context, date time.Time) (*dto.GetScheduleForDayResponse, error)
	GetScheduleRange(ctx context.Context, startDate, endDate time.Time) (*dto.GetScheduleRangeResponse, error)
	GetUpcomingTasks(ctx context.Context, limit int) (*dto.GetUpcomingTasksResponse, error)
	GetStats(ctx context.Context) (*dto.GetStatsResponse, error)
}

type service struct {
	db             *sql.DB
	repo           Repository
	taskRepository goal.TaskRepository
}

func NewService(db *sql.DB, repo Repository, taskRepo goal.TaskRepository) Service {
	return &service{
		db:             db,
		repo:           repo,
		taskRepository: taskRepo,
	}
}

func (s *service) UpdateAvailability(ctx context.Context, goalID uuid.UUID, req dto.UpdateAvailabilityRequest) (*dto.UpdateAvailabilityResponse, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	if err = s.repo.DeleteAvailabilityByGoal(ctx, goalID); err != nil {
		return nil, fmt.Errorf("failed to delete old availability: %w", err)
	}

	now := time.Now()
	for _, dayItem := range req.Days {
		if dayItem.DayOfWeek < 0 || dayItem.DayOfWeek > 6 {
			return nil, fmt.Errorf("invalid day_of_week: %d", dayItem.DayOfWeek)
		}
		av := &Availability{
			ID:        uuid.New(),
			GoalID:    goalID,
			DayOfWeek: dayItem.DayOfWeek,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err = s.repo.CreateAvailability(ctx, av); err != nil {
			return nil, err
		}
		if errSlot := validateTimeSlots(dayItem.Slots); errSlot != nil {
			return nil, errSlot
		}
		for _, slotDTO := range dayItem.Slots {
			st, et, parseErr := parseSlotTimes(slotDTO.StartTime, slotDTO.EndTime)
			if parseErr != nil {
				return nil, parseErr
			}
			slot := &TimeSlot{
				ID:             uuid.New(),
				AvailabilityID: av.ID,
				StartTime:      st,
				EndTime:        et,
				CreatedAt:      now,
				UpdatedAt:      now,
			}
			if err = s.repo.CreateTimeSlot(ctx, slot); err != nil {
				return nil, err
			}
		}
	}

	scheduledCount, err := s.AutoScheduleForGoal(ctx, goalID)
	if err != nil {
		return nil, err
	}

	return &dto.UpdateAvailabilityResponse{
		ScheduledTasks: scheduledCount,
	}, nil
}

func (s *service) ListAvailability(ctx context.Context, goalID uuid.UUID) (*dto.UpdateAvailabilityRequest, error) {
	avList, err := s.repo.ListAvailabilityByGoal(ctx, goalID)
	if err != nil {
		return nil, err
	}
	avIDs := make([]uuid.UUID, 0, len(avList))
	for _, av := range avList {
		avIDs = append(avIDs, av.ID)
	}
	slots, err := s.repo.ListTimeSlotsByAvailabilityIDs(ctx, avIDs)
	if err != nil {
		return nil, err
	}
	avSlotsMap := make(map[uuid.UUID][]TimeSlot)
	for _, slot := range slots {
		avSlotsMap[slot.AvailabilityID] = append(avSlotsMap[slot.AvailabilityID], slot)
	}

	var days []dto.DayAvailability
	for _, av := range avList {
		var slotDTOs []dto.TimeSlotDTO
		for _, sl := range avSlotsMap[av.ID] {
			slotDTOs = append(slotDTOs, dto.TimeSlotDTO{
				StartTime: sl.StartTime.Format("15:04"),
				EndTime:   sl.EndTime.Format("15:04"),
			})
		}
		days = append(days, dto.DayAvailability{
			DayOfWeek: av.DayOfWeek,
			Slots:     slotDTOs,
		})
	}

	return &dto.UpdateAvailabilityRequest{Days: days}, nil
}

func parseSlotTimes(startStr, endStr string) (time.Time, time.Time, error) {
	layout := "15:04"
	st, err1 := time.Parse(layout, startStr)
	et, err2 := time.Parse(layout, endStr)
	if err1 != nil || err2 != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid slot time: %s - %s", startStr, endStr)
	}
	if !st.Before(et) {
		return time.Time{}, time.Time{}, fmt.Errorf("start_time >= end_time: %s - %s", startStr, endStr)
	}
	return st, et, nil
}

func validateTimeSlots(slots []dto.TimeSlotDTO) error {
	// TODO: доделать
	return nil
}

func (s *service) AutoScheduleForGoal(ctx context.Context, goalID uuid.UUID) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *service) GetScheduleForDay(ctx context.Context, date time.Time) (*dto.GetScheduleForDayResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *service) GetScheduleRange(ctx context.Context, startDate, endDate time.Time) (*dto.GetScheduleRangeResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *service) GetUpcomingTasks(ctx context.Context, limit int) (*dto.GetUpcomingTasksResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *service) GetStats(ctx context.Context) (*dto.GetStatsResponse, error) {
	//TODO implement me
	panic("implement me")
}
