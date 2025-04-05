package create

type CreateGoalRequest struct {
	Title         string               `json:"title" validate:"required,max=255"`
	Description   string               `json:"description,omitempty"`
	HoursPerWeek  int                  `json:"hours_per_week" validate:"required,min=1"`
	EstimatedTime int                  `json:"estimated_time"`
	Phases        []CreatePhaseRequest `json:"phases,omitempty"`
}
