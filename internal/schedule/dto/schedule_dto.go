package dto

import "github.com/google/uuid"

type GetScheduleForDayResponse struct {
	Date  string             `json:"date"`
	Tasks []ScheduledTaskDTO `json:"tasks"`
}

type ScheduledTaskDTO struct {
	ID        uuid.UUID `json:"id"`
	GoalTitle string    `json:"goal_title"`
	Title     string    `json:"title"`
	StartTime string    `json:"start_time"`
	EndTime   string    `json:"end_time"`
	Status    string    `json:"status"`
}

type GetScheduleRangeResponse struct {
	Schedule []DaySchedule `json:"schedule"`
}

type DaySchedule struct {
	Date  string             `json:"date"`
	Tasks []ScheduledTaskDTO `json:"tasks"`
}

type GetUpcomingTasksResponse struct {
	Tasks []UpcomingTaskDTO `json:"tasks"`
}

type UpcomingTaskDTO struct {
	ID            uuid.UUID `json:"id"`
	GoalTitle     string    `json:"goal_title"`
	Title         string    `json:"title"`
	ScheduledDate string    `json:"scheduled_date"`
	StartTime     string    `json:"start_time"`
}

type DayProgress struct {
	Day       string `json:"day"`
	Completed int    `json:"completed"`
	Total     int    `json:"total"`
}

type GoalStat struct {
	Title    string `json:"title"`
	Progress int    `json:"progress"`
}

type AutoScheduleResponse struct {
	Message        string `json:"message"`
	ScheduledTasks int    `json:"scheduled_tasks"`
}

type ToggleTaskRequest struct {
	Done bool `json:"done"`
}
