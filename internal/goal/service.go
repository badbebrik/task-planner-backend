package goal

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
	"strings"
	"task-planner/internal/goal/dto"
	"task-planner/internal/goal/dto/create"
	"task-planner/internal/goal/dto/generate"
	"task-planner/internal/goal/dto/get"
	"time"
)

type Service interface {
	CreateGoal(ctx context.Context, userID int64, req create.CreateGoalRequest) (*create.CreateGoalResponse, error)
	GetGoalByID(ctx context.Context, goalID uuid.UUID) (*dto.GoalResponse, error)
	ListGoals(ctx context.Context, userID int64, req get.ListGoalsRequest) (*get.ListGoalsResponse, error)
	GenerateGoalDecomposition(ctx context.Context, userID int64, req generate.GenerateGoalRequest) (*generate.GenerateGoalResponse, error)
	DeleteGoal(ctx context.Context, goalID uuid.UUID) error
	AutoRefillTasks(ctx context.Context, goalID uuid.UUID) (int, error)
}

type service struct {
	repo  RepositoryAggregator
	db    *sql.DB
	aiKey string
}

func NewService(repo RepositoryAggregator, db *sql.DB, openAIKey string) Service {
	return &service{
		repo:  repo,
		db:    db,
		aiKey: openAIKey,
	}
}

func (s *service) CreateGoal(ctx context.Context, userID int64, req create.CreateGoalRequest) (*create.CreateGoalResponse, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	now := time.Now()
	goalID := uuid.New()

	goal := &Goal{
		ID:            goalID,
		UserId:        userID,
		Title:         req.Title,
		Description:   req.Description,
		Status:        "planning",
		EstimatedTime: req.EstimatedTime,
		HoursPerWeek:  req.HoursPerWeek,
		Progress:      0,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err = s.repo.CreateGoal(ctx, goal); err != nil {
		return nil, fmt.Errorf("failed to create goal: %w", err)
	}

	var phaseResponses []dto.PhaseResponse
	for i, phaseReq := range req.Phases {
		phaseID := uuid.New()
		phase := &Phase{
			ID:          phaseID,
			GoalId:      goalID,
			Title:       phaseReq.Title,
			Description: phaseReq.Description,
			Status:      "not_started",
			Progress:    0,
			Order:       phaseReq.Order,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if phase.Order == 0 {
			phase.Order = i + 1
		}
		if err = s.repo.CreatePhase(ctx, phase); err != nil {
			return nil, fmt.Errorf("failed to create phase: %w", err)
		}

		var taskResponses []dto.TaskResponse
		for _, taskReq := range phaseReq.Tasks {
			taskID := uuid.New()
			t := &Task{
				ID:            taskID,
				GoalId:        goalID,
				PhaseId:       &phaseID,
				Title:         taskReq.Title,
				Description:   taskReq.Description,
				Status:        "todo",
				EstimatedTime: taskReq.EstimatedTime,
				CreatedAt:     now,
				UpdatedAt:     now,
			}
			if err = s.repo.CreateTask(ctx, t); err != nil {
				return nil, fmt.Errorf("failed to create task: %w", err)
			}
			taskResponses = append(taskResponses, *s.toTaskResponse(t))
		}

		phResp := s.toPhaseResponse(phase)
		phResp.Tasks = taskResponses

		phaseResponses = append(phaseResponses, *phResp)
	}

	goalResp := s.toGoalResponse(goal)
	goalResp.Phases = phaseResponses

	return &create.CreateGoalResponse{
		Goal: *goalResp,
	}, nil
}

func (s *service) GetGoalByID(ctx context.Context, goalID uuid.UUID) (*dto.GoalResponse, error) {
	g, err := s.repo.GetGoalByID(ctx, goalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get goal: %w", err)
	}
	if g == nil {
		return nil, fmt.Errorf("goal not found")
	}

	phases, err := s.repo.ListPhasesByGoalID(ctx, g.ID)
	if err != nil {
		return nil, err
	}
	tasks, err := s.repo.ListTasksByGoalID(ctx, g.ID)
	if err != nil {
		return nil, err
	}

	for i := range phases {
		ph := &phases[i]
		ph.Progress = ph.CalculateProgress(tasks)
		ph.Status = calculatePhaseStatus(ph)
	}

	_ = s.repo.UpdateGoal(ctx, g)

	goalResp := s.toGoalResponse(g)

	var phaseResponses []dto.PhaseResponse
	for _, p := range phases {
		phResp := s.toPhaseResponse(&p)

		var taskResps []dto.TaskResponse
		for _, t := range tasks {
			if t.PhaseId != nil && *t.PhaseId == p.ID {
				taskResps = append(taskResps, *s.toTaskResponse(&t))
			}
		}
		phResp.Tasks = taskResps
		phaseResponses = append(phaseResponses, *phResp)
	}
	goalResp.Phases = phaseResponses

	return goalResp, nil
}

func (s *service) ListGoals(ctx context.Context, userID int64, req get.ListGoalsRequest) (*get.ListGoalsResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 10
	}

	goals, total, err := s.repo.ListGoals(ctx, userID, req.Limit, req.Offset, req.Status)
	if err != nil {
		return nil, err
	}

	listItems := make([]get.ListGoalItem, 0, len(goals))

	for _, g := range goals {
		phases, err := s.repo.ListPhasesByGoalID(ctx, g.ID)
		if err != nil {
			return nil, err
		}
		tasks, err := s.repo.ListTasksByGoalID(ctx, g.ID)
		if err != nil {
			return nil, err
		}

		for i := range phases {
			p := &phases[i]
			p.Progress = p.CalculateProgress(tasks)
		}

		var nextTask *struct {
			ID      uuid.UUID  `json:"id"`
			Title   string     `json:"title"`
			DueDate *time.Time `json:"due_date,omitempty"`
		}
		for _, t := range tasks {
			if t.Status == "todo" {
				nextTask = (*struct {
					ID      uuid.UUID  `json:"id"`
					Title   string     `json:"title"`
					DueDate *time.Time `json:"due_date,omitempty"`
				})(&struct {
					ID      uuid.UUID
					Title   string
					DueDate *time.Time
				}{
					ID:    t.ID,
					Title: t.Title,
				})
				break
			}
		}

		//_ = s.repo.UpdateGoal(ctx, &g)

		listItems = append(listItems, get.ListGoalItem{
			ID:           g.ID,
			Title:        g.Title,
			Description:  g.Description,
			Status:       g.Status,
			Progress:     g.Progress,
			HoursPerWeek: g.HoursPerWeek,
			UpdatedAt:    g.UpdatedAt,
			NextTask:     nextTask,
		})
	}

	resp := &get.ListGoalsResponse{
		Goals: listItems,
	}
	resp.Meta.Total = total
	resp.Meta.Limit = req.Limit
	resp.Meta.Offset = req.Offset

	return resp, nil
}

func (s *service) GenerateGoalDecomposition(ctx context.Context, userID int64, req generate.GenerateGoalRequest) (*generate.GenerateGoalResponse, error) {
	preview, err := s.callOpenAIForDecomposition(req.Title, req.Description, req.HoursPerWeek)
	if err != nil {
		return nil, err
	}
	return &generate.GenerateGoalResponse{
		GeneratedGoal: *preview,
	}, nil
}

func (s *service) toGoalResponse(g *Goal) *dto.GoalResponse {
	return &dto.GoalResponse{
		ID:            g.ID,
		UserID:        g.UserId,
		Title:         g.Title,
		Description:   g.Description,
		Status:        g.Status,
		HoursPerWeek:  g.HoursPerWeek,
		EstimatedTime: g.EstimatedTime,
		Progress:      g.Progress,
		CreatedAt:     g.CreatedAt,
		UpdatedAt:     g.UpdatedAt,
	}
}

func (s *service) toPhaseResponse(p *Phase) *dto.PhaseResponse {
	return &dto.PhaseResponse{
		ID:          p.ID,
		GoalID:      p.GoalId,
		Title:       p.Title,
		Description: p.Description,
		Status:      p.Status,
		Progress:    p.Progress,
		Order:       p.Order,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func (s *service) toTaskResponse(t *Task) *dto.TaskResponse {
	return &dto.TaskResponse{
		ID:            t.ID,
		GoalID:        t.GoalId,
		PhaseID:       t.PhaseId,
		Title:         t.Title,
		Description:   t.Description,
		Status:        t.Status,
		EstimatedTime: t.EstimatedTime,
		CompletedAt:   t.CompletedAt,
		CreatedAt:     t.CreatedAt,
		UpdatedAt:     t.UpdatedAt,
	}
}

func (s *service) callOpenAIForDecomposition(title, description string, hoursPerWeek int) (*generate.GeneratedGoalPreview, error) {
	prompt := fmt.Sprintf(`
Ты ассистент, помогаешь декомпозировать большие цели на фазы и задачи.
1. Учитывай, что "фаза" – это крупный этап, состоящий из нескольких задач.
2. Задачи – это конкретные, короткие, понятные действия, которые пользователь может выполнить в ближайшее время.
3. В ответе верни JSON со структурой:
{
  "goal": {
    "title": "string",
    "description": "string",
	"hours_per_week": %d,
	"estimated_time": "number",
    "phases": [
      {
        "title": "string",
        "description": "string",
        "estimated_time": "number",
        "tasks": [
          {
            "title": "string",
            "description": "string",
            "estimated_time": "number"
          }
        ]
      },
      ...
    ]
  }
}

ВАЖНЫЕ ПРАВИЛА ДЛЯ ДЕКОМПОЗИЦИИ:
1. Каждая задача (task) должна быть конкретным действием, которое можно выполнить за 1-3 часа
2. Задачи (task) должны быть измеримыми и проверяемыми
3. Сумма времени всех задач в фазе НЕ ДОЛЖНА превышать estimated_time фазы
4. Для первой фазы создавай задачи на первую неделю работы
5. Задачи должны быть последовательными и логически связанными
6. Избегай слишком общих формулировок, используй конкретные действия
7. Каждая задача должна иметь четкий результат

Пример хорошей задачи:
"Создать макеты экранов" - ПЛОХО
"Нарисовать макет главного экрана в Figma" - ХОРОШО

Пример плохой задачи:
"Определить технологии" - ПЛОХО
"Составить список необходимых библиотек для работы с базой данных" - ХОРОШО

Задачи нужны только для первой фазы, для других оставь tasks пустым.

Цель: %s
Описание: %s
Пользователь готов выделять на цель %d часов в неделю`, hoursPerWeek, title, description, hoursPerWeek)

	client := openai.NewClient(s.aiKey)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.7,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}

	type decomposedResult struct {
		Goal generate.GeneratedGoalPreview `json:"goal"`
	}

	var result decomposedResult
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result.Goal, nil
}

func (s *service) DeleteGoal(ctx context.Context, goalID uuid.UUID) error {
	return s.repo.DeleteGoal(ctx, goalID)
}

func calculatePhaseStatus(p *Phase) string {
	progress := p.Progress

	if progress == 0 {
		return "not_started"
	}
	if progress == 100 {
		return "completed"
	}
	return "in_progress"
}

func (s *service) AutoRefillTasks(ctx context.Context, goalID uuid.UUID) (int, error) {
	phase, err := s.currentPhase(ctx, goalID)
	if err != nil || phase == nil {
		return 0, err
	}

	pending, _ := s.repo.CountPendingTasks(ctx, phase.ID)
	if pending >= 3 {
		return 0, nil
	}

	summary, err := s.generatePhaseSummary(ctx, goalID, *phase)
	if err != nil {
		return 0, err
	}

	newTasks, err := s.generateExtraTasks(ctx, goalID, phase, summary, 3-pending)
	if err != nil {
		return 0, err
	}
	if len(newTasks) == 0 {
		return 0, nil
	}

	if err := insertNewTasks(ctx, s.repo, goalID, phase.ID, newTasks); err != nil {
		return 0, err
	}
	return len(newTasks), nil
}

func (s *service) currentPhase(ctx context.Context, goalID uuid.UUID) (*Phase, error) {
	phases, err := s.repo.ListPhasesByGoalID(ctx, goalID)
	if err != nil {
		return nil, err
	}
	var cur *Phase
	for i := range phases {
		if phases[i].Status != "completed" &&
			(cur == nil || phases[i].Order < cur.Order) {
			cur = &phases[i]
		}
	}
	return cur, nil
}

func (s *service) generatePhaseSummary(
	ctx context.Context, goalID uuid.UUID, ph Phase,
) (string, error) {

	since := time.Now().AddDate(0, 0, -14)
	done, _ := s.repo.ListDoneTasksSince(ctx, ph.ID, since)

	var bullets []string
	for _, t := range done {
		bullets = append(bullets, fmt.Sprintf("- %s (%d ч)", t.Title, t.EstimatedTime))
	}

	goalObj, _ := s.repo.GetGoalByID(ctx, goalID)
	prompt := fmt.Sprintf(`
Ты ведёшь дневник выполнения фазы "%s" цели "%s".
За последние 14 дней выполнено:
%s

Опиши краткое саммари (5-7 предложений, какие навыки/темы закрыты, на что делать упор далее).
Ответ JSON:
{ "summary": "..." }`, ph.Title, goalObj.Title, strings.Join(bullets, "\n"))

	resp, err := s.askLLM(ctx, prompt)
	if err != nil {
		return "", err
	}
	var out struct {
		Summary string `json:"summary"`
	}
	if err = json.Unmarshal([]byte(resp), &out); err != nil {
		return "", err
	}
	return out.Summary, nil
}

func (s *service) askLLM(ctx context.Context, prompt string) (string, error) {
	client := openai.NewClient(s.aiKey)
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       openai.GPT4o,
		Temperature: 0.7,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
	})
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}

type llmTask struct {
	Title         string `json:"title"`
	Description   string `json:"description,omitempty"`
	EstimatedTime int    `json:"estimated_time"`
}

func (s *service) generateExtraTasks(
	ctx context.Context,
	goalID uuid.UUID,
	ph *Phase,
	summary string,
	count int,
) ([]llmTask, error) {

	spent, _ := s.repo.SumTimeSpentPhase(ctx, ph.ID)
	remaining := ph.EstimatedTime - spent
	if remaining < 0 {
		remaining = 0
	}
	goal, _ := s.repo.GetGoalByID(ctx, goalID)

	prompt := fmt.Sprintf(`
Ты – ассистент-планировщик. Пользователь работает над фазой "%s" цели "%s".
У него осталось %d из %d запланированных часов фазы.
Он готов уделять цели %d ч/нед.
Последние 2 недели он сделал:
%s

Нужно создать новые конкретные задачи длиной 1-3 ч каждая, 
чтобы продвинуть фазу дальше. Ответ строго в JSON-массиве, 
каждый элемент:
{ "title": "...", "description": "...", "estimated_time": 120 }

Только ASCII кавычки, без комментариев.
ВАЖНЫЕ ПРАВИЛА:
1. Каждая задача должна быть конкретным действием, которое можно выполнить за 1-3 часа
2. Задачи должны быть измеримыми и проверяемыми
3. Сумма времени всех задач в фазе НЕ ДОЛЖНА превышать времени всей фазы
4. Задачи должны быть последовательными и логически связанными
5. Избегай слишком общих формулировок, используй конкретные действия
6. Каждая задача должна иметь четкий результат

Пример хорошей задачи:
"Создать макеты экранов" - ПЛОХО
"Нарисовать макет главного экрана в Figma" - ХОРОШО

Пример плохой задачи:
"Определить технологии" - ПЛОХО
"Составить список необходимых библиотек для работы с базой данных" - ХОРОШО
`,

		ph.Title, goal.Title, remaining, ph.EstimatedTime,
		goal.HoursPerWeek, summary)

	raw, err := s.askLLM(ctx, prompt)
	if err != nil {
		return nil, err
	}
	var out []llmTask
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, fmt.Errorf("bad LLM json: %w, raw: %s", err, raw)
	}
	if len(out) > count {
		out = out[:count]
	}
	return out, nil
}

func insertNewTasks(
	ctx context.Context,
	repo RepositoryAggregator,
	goalID, phaseID uuid.UUID,
	tasks []llmTask,
) error {
	now := time.Now()
	for _, it := range tasks {
		if it.Title == "" || it.EstimatedTime <= 0 {
			continue
		}
		t := &Task{
			ID:            uuid.New(),
			GoalId:        goalID,
			PhaseId:       &phaseID,
			Title:         it.Title,
			Description:   it.Description,
			Status:        "todo",
			EstimatedTime: it.EstimatedTime,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := repo.CreateTask(ctx, t); err != nil {
			return err
		}
	}
	return nil
}
