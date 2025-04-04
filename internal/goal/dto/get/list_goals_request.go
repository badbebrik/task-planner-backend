package get

type ListGoalsRequest struct {
	Limit  int    `json:"limit"  validate:"required,min=1"`
	Offset int    `json:"offset" validate:"required,min=0"`
	Status string `json:"status" validate:"omitempty,oneof=planning active completed paused"`
}
