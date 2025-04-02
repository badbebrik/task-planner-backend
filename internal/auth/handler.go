package auth

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"task-planner/internal/auth/dto"
	"task-planner/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	var req dto.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, ErrInvalidRequest.Error())
		return
	}

	ctx := r.Context()
	err := h.service.Signup(ctx, req.Email, req.Password, req.Name)
	if err != nil {
		if errors.Is(err, ErrUserAlreadyExists) {
			response.Error(w, http.StatusConflict, ErrUserAlreadyExists.Error())
			return
		}

		log.Printf("Failed to register: %v", err)
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(w, http.StatusCreated, "Account successfully created")
}

func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req dto.VerifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, ErrInvalidRequest.Error())
		return
	}

	ctx := r.Context()
	accessToken, refreshToken, err := h.service.VerifyEmailAndGetTokens(ctx, req.Email, req.Code)

	if err != nil {
		log.Printf("Failed to verify emai: %v", err)
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	usr, err := h.service.userService.GetUserByEmail(ctx, req.Email)

	if err != nil {
		response.Error(w, http.StatusNotFound, ErrUserNotFound.Error())
	}

	resp := dto.VerifyEmailResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: dto.UserResponse{
			Email: usr.Email,
			Name:  usr.Name,
			Id:    usr.ID,
		},
	}

	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, ErrInvalidRequest.Error())
		return
	}

	ctx := r.Context()
	accessToken, refreshToken, usr, err := h.service.Login(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			response.Error(w, http.StatusUnauthorized, ErrInvalidCredentials.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: dto.UserResponse{
			Email: usr.Email,
			Name:  usr.Name,
			Id:    usr.ID,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, ErrInvalidRequest.Error())
		return
	}

	ctx := r.Context()
	newAccess, newRefresh, err := h.service.RefreshTokens(ctx, req.RefreshToken)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	resp := dto.RefreshResponse{
		AccessToken:  newAccess,
		RefreshToken: newRefresh,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req dto.LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, ErrInvalidRequest.Error())
		return
	}

	ctx := r.Context()
	if err := h.service.Logout(ctx, req.RefreshToken); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Successfully logged out")
}

func (h *Handler) SendVerificationCode(w http.ResponseWriter, r *http.Request) {
	var req dto.SendVerificationCode
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, ErrInvalidRequest.Error())
		return
	}

	ctx := r.Context()
	err := h.service.SendVerificationCode(ctx, req.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			response.Error(w, http.StatusNotFound, ErrUserNotFound.Error())
			return
		}
		if errors.Is(err, ErrEmailAlreadyVerified) {
			response.Error(w, http.StatusBadRequest, ErrEmailAlreadyVerified.Error())
			return
		}
		log.Printf("Failed to resend verification code: %v", err)
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Verification code sent")
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	claims, err := GetUserFromContext(r.Context())
	if err != nil {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	usr, err := h.service.userService.GetUserByEmail(r.Context(), claims.Email)
	if err != nil {
		http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
		return
	}

	resp := dto.UserResponse{
		Email: usr.Email,
		Name:  usr.Name,
		Id:    usr.ID,
	}

	_ = json.NewEncoder(w).Encode(resp)
}
