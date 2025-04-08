package schedule

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"sort"
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
	if len(slots) <= 1 {
		return nil
	}
	type pair struct{ start, end int }
	var pairs []pair

	for _, s := range slots {
		st, et, err := parseSlotTimes(s.StartTime, s.EndTime)
		if err != nil {
			return err
		}
		startMin := st.Hour()*60 + st.Minute()
		endMin := et.Hour()*60 + et.Minute()
		pairs = append(pairs, pair{startMin, endMin})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].start < pairs[j].start
	})
	for i := 0; i < len(pairs)-1; i++ {
		if pairs[i+1].start < pairs[i].end {
			return errors.New("time slots overlap")
		}
	}
	return nil
}

func (s *service) AutoScheduleForGoal(ctx context.Context, goalID uuid.UUID) (int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	tasks, err := s.taskRepository.ListTasksByGoalID(ctx, goalID)
	if err != nil {
		return 0, fmt.Errorf("list tasks: %w", err)
	}

	var tasksToSchedule []plannedTask
	for _, t := range tasks {
		if t.Status != "todo" {
			continue
		}
		toPlanMinutes := t.EstimatedTime
		if toPlanMinutes <= 0 {
			continue
		}
		tasksToSchedule = append(tasksToSchedule, plannedTask{
			Task:          t,
			RemainingTime: toPlanMinutes,
		})
	}
	if len(tasksToSchedule) == 0 {
		return 0, nil
	}

	avList, err := s.repo.ListAvailabilityByGoal(ctx, goalID)
	if err != nil {
		return 0, fmt.Errorf("list availability: %w", err)
	}
	daySlotsMap, err := s.loadSlotsByDayOfWeek(ctx, avList)
	if err != nil {
		return 0, err
	}

	today := dateOnly(time.Now())
	horizon := 28

	var totalScheduled int

OUTER:
	for dayOffset := 0; dayOffset < horizon; dayOffset++ {
		currentDate := today.AddDate(0, 0, dayOffset)
		dw := int(currentDate.Weekday())
		slotsForDay := daySlotsMap[dw]
		if len(slotsForDay) == 0 {
			continue
		}
		sort.Slice(slotsForDay, func(i, j int) bool {
			return slotsForDay[i].StartTime.Before(slotsForDay[j].StartTime)
		})

		freeIntervals, err := s.calcFreeIntervals(ctx, goalID, currentDate, slotsForDay)
		if err != nil {
			return totalScheduled, err
		}
		if len(freeIntervals) == 0 {
			continue
		}

		for fiIdx, fi := range freeIntervals {
			if fi.duration() <= 0 {
				continue
			}
			for tIdx := range tasksToSchedule {
				if tasksToSchedule[tIdx].RemainingTime <= 0 {
					continue
				}
				needed := tasksToSchedule[tIdx].RemainingTime
				available := fi.duration()

				if available <= 0 {
					break
				}

				if needed <= available {
					st := fi.Start
					end := st.Add(time.Duration(needed) * time.Minute)
					sch := &ScheduledTask{
						ID:            uuid.New(),
						TaskID:        tasksToSchedule[tIdx].Task.ID,
						TimeSlotID:    fi.SlotID,
						ScheduledDate: currentDate,
						StartTime:     st,
						EndTime:       end,
						Status:        "scheduled",
						CreatedAt:     time.Now(),
					}
					if err = s.repo.CreateScheduledTask(ctx, sch); err != nil {
						return totalScheduled, fmt.Errorf("create scheduled task: %w", err)
					}
					totalScheduled++
					tasksToSchedule[tIdx].RemainingTime = 0
					fi.Start = end
				} else {
					st := fi.Start
					end := st.Add(time.Duration(available) * time.Minute)
					sch := &ScheduledTask{
						ID:            uuid.New(),
						TaskID:        tasksToSchedule[tIdx].Task.ID,
						TimeSlotID:    fi.SlotID,
						ScheduledDate: currentDate,
						StartTime:     st,
						EndTime:       end,
						Status:        "scheduled",
						CreatedAt:     time.Now(),
					}
					if err = s.repo.CreateScheduledTask(ctx, sch); err != nil {
						return totalScheduled, err
					}
					totalScheduled++
					tasksToSchedule[tIdx].RemainingTime -= available
					fi.Start = fi.End
				}
				allDone := true
				for _, ptask := range tasksToSchedule {
					if ptask.RemainingTime > 0 {
						allDone = false
						break
					}
				}
				if allDone {
					break OUTER
				}
			}
			freeIntervals[fiIdx] = fi
		}
	}

	return totalScheduled, nil
}

type plannedTask struct {
	Task          goal.Task
	RemainingTime int
}

func (s *service) loadSlotsByDayOfWeek(ctx context.Context, avList []Availability) (map[int][]TimeSlot, error) {
	dayMap := make(map[int][]TimeSlot)
	if len(avList) == 0 {
		return dayMap, nil
	}

	group := make(map[int][]uuid.UUID)
	for _, av := range avList {
		group[av.DayOfWeek] = append(group[av.DayOfWeek], av.ID)
	}
	for dow, avIDs := range group {
		slots, err := s.repo.ListTimeSlotsByAvailabilityIDs(ctx, avIDs)
		if err != nil {
			return nil, err
		}
		dayMap[dow] = append(dayMap[dow], slots...)
	}
	return dayMap, nil
}

func (s *service) calcFreeIntervals(ctx context.Context, goalID uuid.UUID, day time.Time, slots []TimeSlot) ([]freeInterval, error) {
	if len(slots) == 0 {
		return nil, nil
	}
	stList, err := s.repo.ListScheduledTasksForGoalInRange(ctx, goalID, day, day)
	if err != nil {
		return nil, err
	}
	var stSameDay []ScheduledTask
	for _, st := range stList {
		if sameDate(st.ScheduledDate, day) {
			stSameDay = append(stSameDay, st)
		}
	}

	var result []freeInterval
	for _, slot := range slots {
		slotStart := combineDateTime(day, slot.StartTime)
		slotEnd := combineDateTime(day, slot.EndTime)

		var occupied []timeRange
		for _, st := range stSameDay {
			if st.TimeSlotID == slot.ID {
				rng := timeRange{
					start: st.StartTime,
					end:   st.EndTime,
				}
				if rng.start.Before(slotStart) {
					rng.start = slotStart
				}
				if rng.end.After(slotEnd) {
					rng.end = slotEnd
				}
				if rng.start.Before(rng.end) {
					occupied = append(occupied, rng)
				}
			}
		}
		merged := mergeTimeRanges(occupied)
		freeParts := subtractTimeRanges(slotStart, slotEnd, merged)
		for _, f := range freeParts {
			if f.end.After(f.start) {
				result = append(result, freeInterval{
					SlotID: slot.ID,
					Start:  f.start,
					End:    f.end,
				})
			}
		}
	}
	return result, nil
}

type freeInterval struct {
	SlotID uuid.UUID
	Start  time.Time
	End    time.Time
}

func (fi freeInterval) duration() int {
	return int(fi.End.Sub(fi.Start).Minutes())
}

type timeRange struct {
	start time.Time
	end   time.Time
}

func mergeTimeRanges(ranges []timeRange) []timeRange {
	if len(ranges) == 0 {
		return nil
	}
	sort.Slice(ranges, func(i, j int) bool {
		return ranges[i].start.Before(ranges[j].start)
	})
	var merged []timeRange
	current := ranges[0]
	for i := 1; i < len(ranges); i++ {
		nxt := ranges[i]
		if nxt.start.Before(current.end) {
			if nxt.end.After(current.end) {
				current.end = nxt.end
			}
		} else {
			merged = append(merged, current)
			current = nxt
		}
	}
	merged = append(merged, current)
	return merged
}

func subtractTimeRanges(start, end time.Time, busy []timeRange) []timeRange {
	var free []timeRange
	cursor := start
	for _, b := range busy {
		if b.start.After(cursor) {
			free = append(free, timeRange{start: cursor, end: b.start})
		}
		if b.end.After(cursor) {
			cursor = b.end
		}
		if cursor.After(end) {
			break
		}
	}
	if cursor.Before(end) {
		free = append(free, timeRange{start: cursor, end: end})
	}
	return free
}

func sameDate(d1, d2 time.Time) bool {
	return d1.Year() == d2.Year() && d1.Month() == d2.Month() && d1.Day() == d2.Day()
}

func dateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
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
