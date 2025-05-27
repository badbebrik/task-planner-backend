package schedule

import (
	"encoding/json"
	"log"
	"net/http"
	"task-planner/internal/schedule/dto"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"strconv"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// @Summary      Создать или обновить доступность по цели
// @Description  Устанавливает интервалы доступного времени для задач указанной цели и запускает авторасписание
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        goal_id       path      string                                 true  "UUID цели"
// @Param        body          body      dto.UpdateAvailabilityRequest         true  "Данные доступности"
// @Success      200           {object}  dto.UpdateAvailabilityResponse        "Количество запланированных задач"
// @Failure      400           {object}  response.ErrorResponse                      "Invalid goal_id or JSON"
// @Failure      500           {object}  response.ErrorResponse                      "Internal Server Error"
// @Router       /api/availability/{goal_id} [post]
func (h *Handler) CreateOrUpdateAvailability(w http.ResponseWriter, r *http.Request) {
	log.Println("[SCHEDULE] CreateOrUpdateAvailability")

	goalIDStr := chi.URLParam(r, "goal_id")
	goalID, err := uuid.Parse(goalIDStr)
	if err != nil {
		http.Error(w, "Invalid goal_id", http.StatusBadRequest)
		return
	}

	var req dto.UpdateAvailabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	resp, err := h.service.UpdateAvailability(r.Context(), goalID, req)
	if err != nil {
		log.Printf("Error in UpdateAvailability: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// @Summary      Получить доступность по цели
// @Description  Возвращает интервалы доступности для указанной цели
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Produce      json
// @Param        goal_id   path      string                             true  "UUID цели"
// @Success      200       {object}  dto.UpdateAvailabilityRequest     "Данные доступности"
// @Failure      400       {object}  response.ErrorResponse                  "Invalid goal_id"
// @Failure      500       {object}  response.ErrorResponse                  "Internal Server Error"
// @Router       /api/availability/{goal_id} [get]
func (h *Handler) GetAvailability(w http.ResponseWriter, r *http.Request) {
	log.Println("[SCHEDULE] GetAvailability")

	goalIDStr := chi.URLParam(r, "goal_id")
	goalID, err := uuid.Parse(goalIDStr)
	if err != nil {
		http.Error(w, "Invalid goal_id", http.StatusBadRequest)
		return
	}

	resp, err := h.service.ListAvailability(r.Context(), goalID)
	if err != nil {
		log.Printf("Error in ListAvailability: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// @Summary      Авторасписание задач по цели
// @Description  Автоматически планирует "todo"-задачи указанной цели в доступные интервалы
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Produce      json
// @Param        goal_id   path      string  true  "UUID цели"
// @Success      200       {object}  dto.AutoScheduleResponse   "Сообщение и число запланированных задач"
// @Failure      400       {object}  response.ErrorResponse         "Invalid goal_id"
// @Failure      500       {object}  response.ErrorResponse         "Internal Server Error"
// @Router       /api/availability/{goal_id}/schedule [post]
func (h *Handler) AutoSchedule(w http.ResponseWriter, r *http.Request) {
	log.Println("[SCHEDULE] AutoSchedule")
	goalIDStr := chi.URLParam(r, "goal_id")
	goalID, err := uuid.Parse(goalIDStr)
	if err != nil {
		http.Error(w, "Invalid goal_id", http.StatusBadRequest)
		return
	}

	count, err := h.service.AutoScheduleForGoal(r.Context(), goalID)
	if err != nil {
		log.Printf("Error in AutoScheduleForGoal: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"message":         "Auto-schedule complete",
		"scheduled_tasks": count,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// @Summary      Получить расписание
// @Description  Возвращает запланированные задачи за день или диапазон дат
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Produce      json
// @Param        date        query     string  false  "Конкретный день YYYY-MM-DD"
// @Param        start_date  query     string  false  "Начало диапазона YYYY-MM-DD"
// @Param        end_date    query     string  false  "Конец диапазона YYYY-MM-DD"
// @Success      200         {object}  dto.GetScheduleForDayResponse    "Расписание за день"
// @Success      200         {object}  dto.GetScheduleRangeResponse     "Расписание за диапазон"
// @Failure      400         {object}  response.ErrorResponse                "Missing or invalid date parameters"
// @Failure      500         {object}  response.ErrorResponse                "Internal Server Error"
// @Router       /api/schedule [get]
func (h *Handler) GetSchedule(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	if dateStr != "" {
		dt, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			http.Error(w, "Invalid date format (YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
		resp, err := h.service.GetScheduleForDay(r.Context(), dt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	if startDateStr != "" && endDateStr != "" {
		sd, err1 := time.Parse("2006-01-02", startDateStr)
		ed, err2 := time.Parse("2006-01-02", endDateStr)
		if err1 != nil || err2 != nil {
			http.Error(w, "Invalid date format (YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
		resp, err := h.service.GetScheduleRange(r.Context(), sd, ed)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	http.Error(w, "Missing date or start_date/end_date", http.StatusBadRequest)
}

// @Summary      Получить предстоящие задачи
// @Description  Возвращает список ближайших запланированных задач с опциональным лимитом
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Produce      json
// @Param        limit  query     int  false  "Максимальное число задач" default(5)
// @Success      200    {object}  dto.GetUpcomingTasksResponse     "Список задач"
// @Failure      500    {object}  response.ErrorResponse                "Internal Server Error"
// @Router       /api/tasks/upcoming [get]
func (h *Handler) GetUpcomingTasks(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 5
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
			limit = v
		}
	}

	resp, err := h.service.GetUpcomingTasks(r.Context(), limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// @Summary      Получить статистику задач
// @Description  Возвращает статистику выполненных и ожидающих задач за прошлую неделю
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200  {object}  dto.GetStatsResponse   "Статистика по дням"
// @Failure      500  {object}  response.ErrorResponse      "Internal Server Error"
// @Router       /api/stats [get]
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	resp, err := h.service.GetStats(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// @Summary      Отметить или отменить выполнение запланированной задачи
// @Description  Переключает статус запланированной задачи (intervalID) на выполнено или нет
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id    path      string                 true  "UUID запланированного задания"
// @Param        body  body      dto.ToggleTaskRequest  true  "Статус задачи"
// @Success      204   {string}  string                 "No Content"
// @Failure      400   {object}  response.ErrorResponse      "Invalid id or JSON"
// @Failure      500   {object}  response.ErrorResponse      "Internal Server Error"
// @Router       /api/scheduled_tasks/{id} [patch]
func (h *Handler) ToggleInterval(w http.ResponseWriter, r *http.Request) {
	intervalID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var body struct {
		Done bool `json:"done"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if err := h.service.ToggleScheduledTask(r.Context(), intervalID, body.Done); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
