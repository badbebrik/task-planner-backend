package goal

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"log"
	"net/http"
	"task-planner/internal/auth"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateGoal(w http.ResponseWriter, r *http.Request) {
	log.Printf("[GOAL] Starting CreateGoal request")
	
	var req CreateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[GOAL] Failed to decode request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	claims, err := auth.GetUserFromContext(r.Context())
	if err != nil {
		log.Printf("[GOAL] Failed to get user from context: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	goal, err := h.service.CreateGoal(ctx, claims.UserID, req)
	if err != nil {
		log.Printf("[GOAL] Failed to create goal: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Printf("[GOAL] Successfully created goal with ID: %s", goal.ID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(goal)
}

func (h *Handler) GetGoal(w http.ResponseWriter, r *http.Request) {
	log.Printf("[GOAL] Starting GetGoal request")
	
	goalID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("[GOAL] Invalid goal ID: %v", err)
		http.Error(w, "Invalid goal ID", http.StatusBadRequest)
		return
	}

	goal, err := h.service.GetGoal(r.Context(), goalID)
	if err != nil {
		log.Printf("[GOAL] Failed to get goal: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[GOAL] Successfully retrieved goal with ID: %s", goal.ID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(goal)
}

func (h *Handler) UpdateGoal(w http.ResponseWriter, r *http.Request) {
	log.Printf("[GOAL] Starting UpdateGoal request")
	
	goalID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("[GOAL] Invalid goal ID: %v", err)
		http.Error(w, "Invalid goal ID", http.StatusBadRequest)
		return
	}

	var req UpdateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[GOAL] Failed to decode request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	goal, err := h.service.UpdateGoal(r.Context(), goalID, req)
	if err != nil {
		log.Printf("[GOAL] Failed to update goal: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[GOAL] Successfully updated goal with ID: %s", goal.ID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(goal)
}

func (h *Handler) ListGoals(w http.ResponseWriter, r *http.Request) {
	log.Printf("[GOAL] Starting ListGoals request")
	
	var req ListGoalsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[GOAL] Failed to decode request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	claims, err := auth.GetUserFromContext(r.Context())
	if err != nil {
		log.Printf("[GOAL] Failed to get user from context: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	goals, err := h.service.ListGoals(r.Context(), claims.UserID, req)
	if err != nil {
		log.Printf("[GOAL] Failed to list goals: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[GOAL] Successfully retrieved %d goals", len(goals.Goals))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(goals)
}

func (h *Handler) CreatePhase(w http.ResponseWriter, r *http.Request) {
	log.Printf("[GOAL] Starting CreatePhase request")
	
	var req CreatePhaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[GOAL] Failed to decode request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	phase, err := h.service.CreatePhase(r.Context(), req)
	if err != nil {
		log.Printf("[GOAL] Failed to create phase: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[GOAL] Successfully created phase with ID: %s", phase.ID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(phase)
}

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	log.Printf("[GOAL] Starting CreateTask request")
	
	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[GOAL] Failed to decode request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	task, err := h.service.CreateTask(r.Context(), req)
	if err != nil {
		log.Printf("[GOAL] Failed to create task: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[GOAL] Successfully created task with ID: %s", task.ID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (h *Handler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	log.Printf("[GOAL] Starting UpdateTask request")
	
	taskID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("[GOAL] Invalid task ID: %v", err)
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[GOAL] Failed to decode request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	task, err := h.service.UpdateTask(r.Context(), taskID, req)
	if err != nil {
		log.Printf("[GOAL] Failed to update task: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[GOAL] Successfully updated task with ID: %s", task.ID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (h *Handler) CreateGoalDecomposed(w http.ResponseWriter, r *http.Request) {
	log.Printf("[GOAL] Starting CreateGoalDecomposed request")
	
	var req CreateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[GOAL] Failed to decode request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	claims, err := auth.GetUserFromContext(r.Context())
	if err != nil {
		log.Printf("[GOAL] Failed to get user from context: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	goal, err := h.service.CreateGoalDecomposed(r.Context(), claims.UserID, req.Title, req.Description)
	if err != nil {
		log.Printf("[GOAL] Failed to create decomposed goal: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[GOAL] Successfully created decomposed goal with ID: %s", goal.ID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(goal)
}
