package create

type CreatePhaseRequest struct {
	Title         string              `json:"title" validate:"required,max=255"`
	Description   string              `json:"description,omitempty"`
	Order         int                 `json:"order,omitempty"`
	EstimatedTime int                 `json:"estimatedTime"`
	Tasks         []CreateTaskRequest `json:"tasks,omitempty"`
}
