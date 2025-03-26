package goal

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
	"time"
)

type RepositoryAggregator interface {
	GoalRepository
	PhaseRepository
	TaskRepository
}

type Service interface {
	CreateGoal(ctx context.Context, userID int64, req CreateGoalRequest) (*GoalResponse, error)
	GetGoal(ctx context.Context, goalID uuid.UUID) (*GoalResponse, error)
	UpdateGoal(ctx context.Context, goalID uuid.UUID, req UpdateGoalRequest) (*GoalResponse, error)
	ListGoals(ctx context.Context, userID int64, req ListGoalsRequest) (*ListGoalsResponse, error)
	CreatePhase(ctx context.Context, req CreatePhaseRequest) (*PhaseResponse, error)
	CreateTask(ctx context.Context, req CreateTaskRequest) (*TaskResponse, error)
	UpdateTask(ctx context.Context, taskID uuid.UUID, req UpdateTaskRequest) (*TaskResponse, error)
	CreateGoalDecomposed(ctx context.Context, userID int64, title, description string) (*GoalResponse, error)
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

func (s *service) CreateGoal(ctx context.Context, userID int64, req CreateGoalRequest) (*GoalResponse, error) {
	goalID := uuid.New()
	g := &Goal{
		ID:            goalID,
		UserId:        userID,
		Title:         req.Title,
		Description:   req.Description,
		Status:        "in-progress",
		EstimatedTime: 0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repo.CreateGoal(ctx, g); err != nil {
		return nil, fmt.Errorf("failed to create goal: %w", err)
	}

	return s.toGoalResponse(g), nil
}

func (s *service) GetGoal(ctx context.Context, goalID uuid.UUID) (*GoalResponse, error) {
	g, err := s.repo.GetGoalByID(ctx, goalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get goal: %w", err)
	}
	if g == nil {
		return nil, fmt.Errorf("goal not found")
	}

	return s.toGoalResponse(g), nil
}

func (s *service) UpdateGoal(ctx context.Context, goalID uuid.UUID, req UpdateGoalRequest) (*GoalResponse, error) {
	g, err := s.repo.GetGoalByID(ctx, goalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get goal: %w", err)
	}
	if g == nil {
		return nil, fmt.Errorf("goal not found")
	}

	if req.Title != "" {
		g.Title = req.Title
	}
	if req.Description != "" {
		g.Description = req.Description
	}
	if req.Status != "" {
		g.Status = req.Status
	}

	if err := s.repo.UpdateGoal(ctx, g); err != nil {
		return nil, fmt.Errorf("failed to update goal: %w", err)
	}

	return s.toGoalResponse(g), nil
}

func (s *service) ListGoals(ctx context.Context, userID int64, req ListGoalsRequest) (*ListGoalsResponse, error) {
	// TODO: Implement pagination, filtering and sorting
	return nil, fmt.Errorf("not implemented")
}

func (s *service) CreatePhase(ctx context.Context, req CreatePhaseRequest) (*PhaseResponse, error) {
	phaseID := uuid.New()
	p := &Phase{
		ID:            phaseID,
		GoalId:        req.GoalID,
		Title:         req.Title,
		Description:   req.Description,
		Status:        "in-progress",
		EstimatedTime: req.EstimatedTime,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repo.CreatePhase(ctx, p); err != nil {
		return nil, fmt.Errorf("failed to create phase: %w", err)
	}

	return s.toPhaseResponse(p), nil
}

func (s *service) CreateTask(ctx context.Context, req CreateTaskRequest) (*TaskResponse, error) {
	taskID := uuid.New()
	t := &Task{
		ID:            taskID,
		GoalId:        req.GoalID,
		PhaseId:       req.PhaseID,
		Title:         req.Title,
		Description:   req.Description,
		Status:        "in-progress",
		EstimatedTime: req.EstimatedTime,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repo.CreateTask(ctx, t); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return s.toTaskResponse(t), nil
}

func (s *service) UpdateTask(ctx context.Context, taskID uuid.UUID, req UpdateTaskRequest) (*TaskResponse, error) {
	// TODO: Implement task update
	return nil, fmt.Errorf("not implemented")
}

func (s *service) CreateGoalDecomposed(ctx context.Context, userID int64, title, description string) (*GoalResponse, error) {
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

	return s.toGoalResponse(g), nil
}

func (s *service) toGoalResponse(g *Goal) *GoalResponse {
	return &GoalResponse{
		ID:            g.ID,
		UserID:        g.UserId,
		Title:         g.Title,
		Description:   g.Description,
		Status:        g.Status,
		EstimatedTime: g.EstimatedTime,
		CreatedAt:     g.CreatedAt,
		UpdatedAt:     g.UpdatedAt,
	}
}

func (s *service) toPhaseResponse(p *Phase) *PhaseResponse {
	return &PhaseResponse{
		ID:            p.ID,
		GoalID:        p.GoalId,
		Title:         p.Title,
		Description:   p.Description,
		Status:        p.Status,
		EstimatedTime: p.EstimatedTime,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

func (s *service) toTaskResponse(t *Task) *TaskResponse {
	return &TaskResponse{
		ID:            t.ID,
		GoalID:        t.GoalId,
		PhaseID:       t.PhaseId,
		Title:         t.Title,
		Description:   t.Description,
		Status:        t.Status,
		EstimatedTime: t.EstimatedTime,
		CreatedAt:     t.CreatedAt,
		UpdatedAt:     t.UpdatedAt,
	}
}

type decompositionResult struct {
	Phases []Phase `json:"phases"`
	Tasks  []Task  `json:"tasks"`
}

func (s *service) callOpenAIForDecomposition(title, description string) ([]Phase, []Task, error) {
	prompt := fmt.Sprintf(`
Ты ассистент, помогаешь декомпозировать большие цели на фазы и задачи.
1. Учитывай, что "фаза" – это крупный этап, состоящий из нескольких задач.
2. Задачи – это конкретные, короткие, понятные действия, которые пользователь может выполнить в ближайшее время.
3. В ответе верни JSON со структурой:
{
  "phases": [
    {
      "title": "string",
      "description": "string",
      "estimated_time": "number"
    },
    ...
  ],
  "tasks": [
    {
      "title": "string",
      "description": "string"
      "estimated_time": "number"
    },
    ...
  ]
}

Важно: кроме этого JSON ничего не пиши, никаких слов или пояснений вне JSON. EstimatedTime указывать в часах

Цель: %s
Описание: %s
`, title, description)
	client := openai.NewClient(s.aiKey)
	req := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "Ты – специалист по планированию целей. Отвечаешь строго в JSON.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.2,
	}
	resp, err := client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return nil, nil, fmt.Errorf("openAI CreateChatCompletion error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, nil, fmt.Errorf("no choices in openAI response")
	}

	raw := resp.Choices[0].Message.Content
	var result decompositionResult
	err = json.Unmarshal([]byte(raw), &result)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal openAI response: %w\nOpenAI response was: %s", err, raw)
	}
	return result.Phases, result.Tasks, nil
}
