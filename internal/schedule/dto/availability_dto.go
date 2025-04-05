package dto

type UpdateAvailabilityRequest struct {
	Days []DayAvailability `json:"days"`
}

type DayAvailability struct {
	DayOfWeek int           `json:"day_of_week"`
	Slots     []TimeSlotDTO `json:"slots"`
}

type TimeSlotDTO struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type UpdateAvailabilityResponse struct {
	ScheduledTasks int `json:"scheduled_tasks"`
}
