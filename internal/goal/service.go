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
	CreateGoalDecomposed(ctx context.Context, userID int64, title, description string) (*Goal, []Phase, []Task, error)
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

func (s *service) CreateGoalDecomposed(ctx context.Context, userID int64, title, description string) (*Goal, []Phase, []Task, error) {
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
		return nil, nil, nil, fmt.Errorf("failed to call openAI: %w", err)
	}

	for _, p := range phases {
		p.ID = uuid.New()
		p.GoalId = goalID
		p.Status = "in-progress"
		p.CreatedAt = time.Now()
		p.UpdatedAt = time.Now()
	}

	for _, t := range tasks {
		t.ID = uuid.New()
		t.GoalId = goalID
		t.Status = "in-progress"
		t.CreatedAt = time.Now()
		t.UpdatedAt = time.Now()
	}

	return g, phases, tasks, nil
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
