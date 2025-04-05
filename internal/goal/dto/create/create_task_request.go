package create

type CreateTaskRequest struct {
	Title         string `json:"title" validate:"required,max=255"`
	Description   string `json:"description,omitempty"`
	EstimatedTime int    `json:"estimated_time"`
}
