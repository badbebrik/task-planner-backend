package create

type CreatePhaseRequest struct {
	Title       string              `json:"title" validate:"required,max=255"`
	Description string              `json:"description,omitempty"`
	Order       int                 `json:"order,omitempty"`
	Tasks       []CreateTaskRequest `json:"tasks,omitempty"`
}
