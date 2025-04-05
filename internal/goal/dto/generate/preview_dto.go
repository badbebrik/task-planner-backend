package generate

type GeneratedGoalPreview struct {
	Title         string                `json:"title"`
	Description   string                `json:"description,omitempty"`
	HoursPerWeek  int                   `json:"hours_per_week"`
	EstimatedTime int                   `json:"estimated_time"`
	Phases        []GeneratedPhaseDraft `json:"phases"`
}

type GeneratedPhaseDraft struct {
	Title         string               `json:"title"`
	Description   string               `json:"description,omitempty"`
	EstimatedTime int                  `json:"estimatedTime"`
	Order         int                  `json:"order"`
	Tasks         []GeneratedTaskDraft `json:"tasks"`
}

type GeneratedTaskDraft struct {
	Title         string `json:"title"`
	Description   string `json:"description,omitempty"`
	EstimatedTime int    `json:"estimated_time"`
}
