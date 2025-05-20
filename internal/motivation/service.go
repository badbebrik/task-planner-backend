package motivation

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
	"strings"
	"task-planner/internal/goal"
	"time"
)

type Service interface {
	GenerateDailyMotivations(ctx context.Context) error
	GetTodayMotivation(ctx context.Context, userID int64) (string, error)
}

type service struct {
	repo           Repository
	taskRepository goal.TaskRepository
	aiKey          string
}

func NewService(repo Repository, taskRepo goal.TaskRepository, openAIKey string) Service {
	return &service{repo: repo, taskRepository: taskRepo, aiKey: openAIKey}
}

func (s *service) GenerateDailyMotivations(ctx context.Context) error {
	today := dateOnly(time.Now())

	users, err := s.taskRepository.ListUsersWithTasksOnDate(ctx, today)
	if err != nil {
		return err
	}

	for _, userID := range users {
		existing, err := s.repo.GetByUserAndDate(ctx, userID, today)
		if err != nil {
			return err
		}
		if existing != nil {
			continue
		}

		tasks, err := s.taskRepository.ListTasksByUserAndDate(ctx, userID, today)
		if err != nil {
			return err
		}

		prompt := buildMotivationPrompt(tasks)

		client := openai.NewClient(s.aiKey)
		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: openai.GPT4,
				Messages: []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleUser, Content: prompt},
				},
				Temperature: 0.8,
			},
		)
		if err != nil {
			return fmt.Errorf("motivation LLM: %w", err)
		}
		text := resp.Choices[0].Message.Content

		m := &Motivation{
			ID:        uuid.New(),
			UserID:    userID,
			Date:      today,
			Text:      text,
			CreatedAt: time.Now(),
		}
		if err := s.repo.Create(ctx, m); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) GetTodayMotivation(ctx context.Context, userID int64) (string, error) {
	today := dateOnly(time.Now())
	m, err := s.repo.GetByUserAndDate(ctx, userID, today)
	if err != nil {
		return "", err
	}
	if m != nil {
		return m.Text, nil
	}

	if err := s.GenerateDailyMotivations(ctx); err != nil {
		return "", err
	}
	m, err = s.repo.GetByUserAndDate(ctx, userID, today)
	if err != nil {
		return "", err
	}
	if m != nil {
		return m.Text, nil
	}
	return "Желаем отличного дня!", nil
}

func dateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func buildMotivationPrompt(tasks []goal.Task) string {
	if len(tasks) == 0 {
		return "Подбодри меня и скажи что-нибудь вдохновляющее для отдыха или похвалы."
	}
	var lines []string
	for _, t := range tasks {
		lines = append(lines, fmt.Sprintf("- %s", t.Title))
	}
	return fmt.Sprintf(`
У меня сегодня задачи:
%s

Напиши короткое вдохновляющее сообщение, чтобы я замотивировался выполнить именно эти дела.
`, strings.Join(lines, "\n"))
}
