package goal

import (
	"encoding/json"
	"log"
	"net/http"
	"task-planner/internal/goal/dto/create"
	"task-planner/internal/goal/dto/generate"
	"task-planner/internal/goal/dto/get"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"task-planner/internal/auth"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// @Summary      Генерация разбивки цели
// @Description  Создаёт рекомендуемую декомпозицию цели на фазы и задачи через LLM
// @Tags         Goal
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        GenerateGoalRequest  body      generate.GenerateGoalRequest  true  "Данные для генерации цели"
// @Success      200                  {object}  generate.GenerateGoalResponse
// @Failure      400                  {object}  response.ErrorResponse       "Invalid request body"
// @Failure      401                  {object}  response.ErrorResponse       "Unauthorized"
// @Failure      500                  {object}  response.ErrorResponse       "Failed to generate goal"
// @Router       /api/goals/generate [post]
func (h *Handler) GenerateGoal(w http.ResponseWriter, r *http.Request) {
	log.Println("[GOAL] GenerateGoal request")

	var req generate.GenerateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	claims, err := auth.GetUserFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	resp, err := h.service.GenerateGoalDecomposition(r.Context(), claims.UserID, req)
	if err != nil {
		log.Printf("[GOAL] Failed to generate: %v", err)
		http.Error(w, "Failed to generate goal", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// @Summary      Создание цели
// @Description  Сохраняет новую цель вместе с фазами и задачами в базе
// @Tags         Goal
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        CreateGoalRequest  body      create.CreateGoalRequest  true  "Данные новой цели"
// @Success      201                {object}  create.CreateGoalResponse
// @Failure      400                {object}  response.ErrorResponse  "Invalid request body"
// @Failure      401                {object}  response.ErrorResponse  "Unauthorized"
// @Failure      500                {object}  response.ErrorResponse  "Internal Server Error"
// @Router       /api/goals [post]
func (h *Handler) CreateGoal(w http.ResponseWriter, r *http.Request) {
	log.Println("[GOAL] CreateFullGoal request")

	var req create.CreateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	claims, err := auth.GetUserFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	result, err := h.service.CreateGoal(r.Context(), claims.UserID, req)
	if err != nil {
		log.Printf("[GOAL] failed to create full goal: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// @Summary      Список целей
// @Description  Возвращает постраничный список целей пользователя с фильтром по статусу
// @Tags         Goal
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        status  query     string  false  "Фильтр по статусу (planning,in_progress,completed)"
// @Param        limit   query     int     false  "Максимальное число элементов" default(10)
// @Param        offset  query     int     false  "Смещение для пагинации" default(0)
// @Success      200     {object}  get.ListGoalsResponse
// @Failure      401     {object}  response.ErrorResponse  "Unauthorized"
// @Failure      500     {object}  response.ErrorResponse  "Internal Server Error"
// @Router       /api/goals [get]
func (h *Handler) ListGoals(w http.ResponseWriter, r *http.Request) {
	log.Println("[GOAL] ListGoals request")

	claims, err := auth.GetUserFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	limit := 10
	offset := 0
	status := r.URL.Query().Get("status")

	reqStruct := get.ListGoalsRequest{
		Limit:  limit,
		Offset: offset,
		Status: status,
	}

	resp, err := h.service.ListGoals(r.Context(), claims.UserID, reqStruct)
	if err != nil {
		log.Printf("[GOAL] failed to list goals: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// @Summary      Получить цель по ID
// @Description  Возвращает подробную информацию о цели, включая фазы и задачи
// @Tags         Goal
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id   path      string  true  "UUID цели"
// @Success      200  {object}  dto.GoalResponse  "Подробная информация о цели"
// @Failure      400  {object}  response.ErrorResponse  "Invalid goal ID"
// @Failure      404  {object}  response.ErrorResponse  "Goal not found"
// @Failure      500  {object}  response.ErrorResponse  "Internal Server Error"
// @Router       /api/goals/{id} [get]
func (h *Handler) GetGoal(w http.ResponseWriter, r *http.Request) {
	log.Println("[GOAL] GetGoal request")

	goalIDStr := chi.URLParam(r, "id")
	goalID, err := uuid.Parse(goalIDStr)
	if err != nil {
		http.Error(w, "Invalid goal ID", http.StatusBadRequest)
		return
	}

	goalResp, err := h.service.GetGoalByID(r.Context(), goalID)
	if err != nil {
		log.Printf("[GOAL] failed to get goal: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if goalResp == nil {
		http.Error(w, "Goal not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"goal": goalResp,
	})
}

// @Summary      Удалить цель
// @Description  Удаляет цель по ID
// @Tags         Goal
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id   path      string  true  "UUID цели"
// @Success      204  {string}  string  "No Content"
// @Failure      400  {object}  response.ErrorResponse  "Invalid goal ID"
// @Failure      401  {object}  response.ErrorResponse  "Unauthorized"
// @Failure      500  {object}  response.ErrorResponse  "Internal Server Error"
// @Router       /api/goals/{id} [delete]
func (h *Handler) DeleteGoal(w http.ResponseWriter, r *http.Request) {
	goalID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid goal ID", http.StatusBadRequest)
		return
	}

	if _, err := auth.GetUserFromContext(r.Context()); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.service.DeleteGoal(r.Context(), goalID); err != nil {
		log.Printf("[GOAL] delete failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
