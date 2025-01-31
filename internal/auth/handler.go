package auth

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type RegisterEmailRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterEmail(w http.ResponseWriter, r *http.Request) {
	var req RegisterEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	err := h.service.RegisterEmail(ctx, req.Email, req.Password, req.Name)
	if err != nil {
		if errors.Is(err, ErrUserAlreadyExists) {
			http.Error(w, "User already exist", http.StatusConflict)
			return
		}
		log.Printf("Failed to register email: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message":"Verification email sent"}`))
}

type VerifyEmailRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req VerifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	err := h.service.VerifyEmail(ctx, req.Email, req.Code)

	if err != nil {
		log.Printf("Failed to verify emai: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Email verified successfully"}`))
}
