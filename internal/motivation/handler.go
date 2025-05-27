package motivation

import (
	"encoding/json"
	"net/http"
	"task-planner/internal/auth"
)

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

// @Summary      Получить мотивацию на сегодня
// @Description  Возвращает текст мотивации для текущего пользователя
// @Tags         Motivation
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  map[string]string  "{\"motivation\": \"...\"}"
// @Failure      401  {object}  response.ErrorResponse  "Unauthorized"
// @Failure      500  {object}  response.ErrorResponse  "Internal Server Error"
// @Router       /api/motivation/today [get]
func (h *Handler) GetToday(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.GetUserFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	text, err := h.service.GetTodayMotivation(r.Context(), claims.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"motivation": text})
}
