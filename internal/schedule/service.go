package schedule

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"log"
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
	log.Printf("[AutoSchedule] total tasks from repo: %d", len(tasks))
	if err != nil {
		return 0, fmt.Errorf("list tasks: %w", err)
	}

	var tasksToSchedule []plannedTask
	for _, t := range tasks {
		log.Printf("[AutoSchedule] task %s status=%q est=%d", t.ID, t.Status, t.EstimatedTime)
		if t.Status != "todo" {
			continue
		}
		toPlanMinutes := t.EstimatedTime * 60
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
	log.Printf("[AutoSchedule] availability for goal %s: %+v", goalID, avList)

	if err != nil {
		return 0, fmt.Errorf("list availability: %w", err)
	}
	daySlotsMap, err := s.loadSlotsByDayOfWeek(ctx, avList)
	for dow, slots := range daySlotsMap {
		log.Printf("[AutoSchedule] dow=%d slotsCount=%d", dow, len(slots))
		for _, slot := range slots {
			log.Printf("  [Slot] id=%s start=%s end=%s",
				slot.ID,
				slot.StartTime.Format("15:04"),
				slot.EndTime.Format("15:04"))
		}
	}

	if err != nil {
		return 0, err
	}

	today := dateOnly(time.Now())
	horizon := 28

	var totalScheduled int

OUTER:
	for dayOffset := 0; dayOffset < horizon; dayOffset++ {
		currentDate := today.AddDate(0, 0, dayOffset)
		dow := int(currentDate.Weekday())
		slotsForDay := daySlotsMap[dow]

		log.Printf("[AutoSchedule] checking date %s (dow=%d), slotsForDay=%d",
			currentDate.Format("2006-01-02"), dow, len(slotsForDay))

		if len(slotsForDay) == 0 {
			continue
		}
		sort.Slice(slotsForDay, func(i, j int) bool {
			return slotsForDay[i].StartTime.Before(slotsForDay[j].StartTime)
		})

		freeIntervals, err := s.calcFreeIntervals(ctx, goalID, currentDate, slotsForDay)
		log.Printf("[AutoSchedule][%s] freeIntervals: %+v", currentDate.Format("2006-01-02"), freeIntervals)
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

	log.Printf("[AutoSchedule] finished, totalScheduled=%d", totalScheduled)

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
	log.Printf("[GetScheduleForDay] called with date=%s", date.Format("2006-01-02"))
	scheduledList, err := s.repo.ListScheduledTasksInRange(ctx, date, date)
	if err != nil {
		return nil, fmt.Errorf("failed to list scheduled tasks for day: %w", err)
	}
	log.Printf("[GetScheduleForDay] scheduledList.len=%d items=%+v",
		len(scheduledList), scheduledList)

	tasksMap, goalsMap, err := s.loadTasksAndGoals(ctx, scheduledList)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks/goals: %w", err)
	}

	items := make([]dto.ScheduledTaskDTO, 0)
	for _, st := range scheduledList {
		log.Printf("[GetScheduleForDay] processing st.ID=%s, st.TaskID=%s, st.StartTime=%s, st.EndTime=%s",
			st.ID, st.TaskID, st.StartTime.Format("15:04"), st.EndTime.Format("15:04"))
		t := tasksMap[st.TaskID]
		g := goalsMap[t.GoalId]

		items = append(items, dto.ScheduledTaskDTO{
			ID:        st.ID,
			GoalTitle: g.Title,
			Title:     t.Title,
			StartTime: st.StartTime.Format("15:04"),
			EndTime:   st.EndTime.Format("15:04"),
			Status:    st.Status,
		})
	}

	log.Printf("[GetScheduleForDay] returning %d tasks", len(items))
	resp := &dto.GetScheduleForDayResponse{
		Date:  date.Format("2006-01-02"),
		Tasks: items,
	}
	return resp, nil
}

func (s *service) GetScheduleRange(ctx context.Context, startDate, endDate time.Time) (*dto.GetScheduleRangeResponse, error) {
	scheduledList, err := s.repo.ListScheduledTasksInRange(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to list scheduled tasks in range: %w", err)
	}

	tasksMap, goalsMap, err := s.loadTasksAndGoals(ctx, scheduledList)
	if err != nil {
		return nil, err
	}

	grouped := make(map[string][]dto.ScheduledTaskDTO)
	for _, st := range scheduledList {
		dateKey := st.ScheduledDate.Format("2006-01-02")

		t := tasksMap[st.TaskID]
		g := goalsMap[t.GoalId]

		grouped[dateKey] = append(grouped[dateKey], dto.ScheduledTaskDTO{
			ID:        st.ID,
			GoalTitle: g.Title,
			Title:     t.Title,
			StartTime: st.StartTime.Format("15:04"),
			EndTime:   st.EndTime.Format("15:04"),
			Status:    st.Status,
		})
	}

	scheduleResult := make([]dto.DaySchedule, 0)
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateKey := d.Format("2006-01-02")
		tasks := grouped[dateKey]
		scheduleResult = append(scheduleResult, dto.DaySchedule{
			Date:  dateKey,
			Tasks: tasks,
		})
	}

	return &dto.GetScheduleRangeResponse{
		Schedule: scheduleResult,
	}, nil
}

func (s *service) GetUpcomingTasks(ctx context.Context, limit int) (*dto.GetUpcomingTasksResponse, error) {
	if limit <= 0 {
		limit = 5
	}
	stList, err := s.repo.ListUpcomingTasks(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list upcoming tasks: %w", err)
	}

	tasksMap, goalsMap, err := s.loadTasksAndGoals(ctx, stList)
	if err != nil {
		return nil, err
	}

	var items []dto.UpcomingTaskDTO
	for _, st := range stList {
		t := tasksMap[st.TaskID]
		g := goalsMap[t.GoalId]
		items = append(items, dto.UpcomingTaskDTO{
			ID:            st.ID,
			GoalTitle:     g.Title,
			Title:         t.Title,
			ScheduledDate: st.ScheduledDate.Format("2006-01-02"),
			StartTime:     st.StartTime.Format("15:04"),
		})
	}
	return &dto.GetUpcomingTasksResponse{Tasks: items}, nil
}

func (s *service) GetStats(ctx context.Context) (*dto.GetStatsResponse, error) {
	today := dateOnly(time.Now())
	startDate := today.AddDate(0, 0, -6)

	dayMap, err := s.repo.CountTasksByDay(ctx, startDate, today)
	if err != nil {
		return nil, fmt.Errorf("failed to count tasks by day: %w", err)
	}

	var weekly []dto.DayProgress
	for d := startDate; !d.After(today); d = d.AddDate(0, 0, 1) {
		data, ok := dayMap[d]
		wdayStr := d.Weekday().String()[:3]
		if !ok {
			weekly = append(weekly, dto.DayProgress{
				Day:       wdayStr,
				Completed: 0,
				Total:     0,
			})
			continue
		}
		weekly = append(weekly, dto.DayProgress{
			Day:       wdayStr,
			Completed: data.Completed,
			Total:     data.Total,
		})
	}

	goals := []dto.GoalStat{
		{Title: "Выучить испанский", Progress: 35},
		{Title: "Написать приложение", Progress: 10},
	}

	return &dto.GetStatsResponse{
		WeeklyProgress: weekly,
		Goals:          goals,
	}, nil
}

func (s *service) loadTasksAndGoals(
	ctx context.Context,
	scheduledList []ScheduledTask,
) (
	map[uuid.UUID]goal.Task,
	map[uuid.UUID]goal.Goal,
	error,
) {
	taskIDs := make(map[uuid.UUID]struct{})
	for _, st := range scheduledList {
		taskIDs[st.TaskID] = struct{}{}
	}

	var ids []uuid.UUID
	for id := range taskIDs {
		ids = append(ids, id)
	}

	tasks, err := s.taskRepository.GetTasksByIDs(ctx, ids)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	taskMap := make(map[uuid.UUID]goal.Task)
	goalIDs := make(map[uuid.UUID]struct{})
	for _, t := range tasks {
		taskMap[t.ID] = t
		goalIDs[t.GoalId] = struct{}{}
	}

	var gIDs []uuid.UUID
	for id := range goalIDs {
		gIDs = append(gIDs, id)
	}

	goals, err := s.taskRepository.GetGoalsByIDs(ctx, gIDs)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load goals: %w", err)
	}

	goalMap := make(map[uuid.UUID]goal.Goal)
	for _, g := range goals {
		goalMap[g.ID] = g
	}

	return taskMap, goalMap, nil
}
