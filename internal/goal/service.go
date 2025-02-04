package goal

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"time"
)

type RepositoryAggregator interface {
	GoalRepository
	PhaseRepository
	TaskRepository
}

type Service interface {
	CreateGoalDecomposed(ctx context.Context, userID int64, title, description string) (*Goal, error)
}

type service struct {
	repo  RepositoryAggregator
	aiKey string
}

func NewService(repo RepositoryAggregator, openAIKey string) Service {
	return &service{
		repo:  repo,
		aiKey: openAIKey,
	}
}

func (s *service) CreateGoalDecomposed(ctx context.Context, userID int64, title, description string) (*Goal, error) {
	goalID := uuid.New()
	g := &Goal{
		ID:            goalID,
		UserId:        userID,
		Title:         title,
		Description:   description,
		Status:        "in-progress",
		EstimatedTime: 0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repo.CreateGoal(ctx, g); err != nil {
		return nil, fmt.Errorf("failed to create goal: %w", err)
	}

	phases, tasks, err := s.callOpenAIForDecomposition(title, description)
	if err != nil {
		return nil, fmt.Errorf("failed to call openAI: %w", err)
	}

	for _, p := range phases {
		p.ID = uuid.New()
		p.GoalId = goalID
		p.Status = "in-progress"
		p.CreatedAt = time.Now()
		p.UpdatedAt = time.Now()

		if err := s.repo.CreatePhase(ctx, &p); err != nil {
			return nil, fmt.Errorf("failed to create phase: %w", err)
		}
	}

	for _, t := range tasks {
		t.ID = uuid.New()
		t.GoalId = goalID
		t.Status = "in-progress"
		t.CreatedAt = time.Now()
		t.UpdatedAt = time.Now()

		if err := s.repo.CreateTask(ctx, &t); err != nil {
			return nil, fmt.Errorf("failed to create task: %w", err)
		}
	}

	return g, nil
}

func (s *service) callOpenAIForDecomposition(title, description string) ([]Phase, []Task, error) {
	prompt := fmt.Sprintf(```Я хочу разбить цель на фазы и задачи. Цель: '%s'. Описание: '%s'.
	Опиши мне несколько фаз (2-5 штук), и для каждой фазы по 1-3 задачи.
		Каждая задача или фаза должны включать 'title', 'description', 'estimated_time'.
		Верни результат строго в формате JSON вида:
	{
		\"phases\": [
		{
			\"title\": \"\",
			\"description\": \"\",
			\"estimated_time\": 0
		}
],
	\"tasks\": [
	{
	\"title\": \"\",
	\"description\": \"\",
	\"estimated_time\": 0,
	\"phase_title\": \"\" 
	}
]
}```, title, description)

}
