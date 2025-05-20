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
