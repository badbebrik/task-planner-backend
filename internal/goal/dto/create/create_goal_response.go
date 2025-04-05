package create

import "task-planner/internal/goal/dto"

type CreateGoalResponse struct {
	Goal dto.GoalResponse `json:"goal"`
}
