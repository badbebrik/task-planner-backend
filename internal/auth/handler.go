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

// @Summary      Регистрация пользователя
// @Description  Регистрирует нового пользователя и отправляет код верификации на email
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        signupRequest  body      dto.SignupRequest  true  "Данные регистрации"
// @Success      201            {string}  string             "Account successfully created"
// @Failure      400            {object}  response.ErrorResponse
// @Router       /api/auth/signup [post]
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

// @Summary      Подтверждение email
// @Description  Проверяет код подтверждения и возвращает JWT-токены + данные пользователя
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        verifyEmailRequest  body      dto.VerifyEmailRequest  true  "Email и код верификации"
// @Success      200                 {object}  dto.VerifyEmailResponse
// @Failure      400                 {object}  response.ErrorResponse
// @Failure      404                 {object}  response.ErrorResponse
// @Router       /api/auth/verify-email [post]
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

// @Summary      Вход пользователя
// @Description  Аутентифицирует по email и паролю, возвращает JWT-токены + данные пользователя
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        loginRequest  body      dto.LoginRequest  true  "Данные для входа"
// @Success      200           {object}  dto.LoginResponse
// @Failure      400           {object}  response.ErrorResponse
// @Failure      401           {object}  response.ErrorResponse
// @Router       /api/auth/login [post]
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

// @Summary      Обновление токенов
// @Description  Обменивает refresh-токен на новую пару access/refresh
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        refreshRequest  body      dto.RefreshRequest  true  "Refresh-токен"
// @Success      200             {object}  dto.RefreshResponse
// @Failure      400             {object}  response.ErrorResponse
// @Failure      401             {object}  response.ErrorResponse
// @Router       /api/auth/refresh [post]
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

// @Summary      Выход (logout)
// @Description  Ревокирует переданный refresh-токен
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        logoutRequest  body      dto.LogoutRequest  true  "Refresh-токен"
// @Success      200           {string}  string             "Successfully logged out"
// @Failure      400           {object}  response.ErrorResponse
// @Router       /api/auth/logout [post]
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

// @Summary      Повторная отправка кода верификации
// @Description  Отправляет новый код подтверждения на email пользователя
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        sendVerificationCode  body      dto.SendVerificationCode  true  "Email пользователя"
// @Success      200                   {string}  string                   "Verification code sent"
// @Failure      400                   {object}  response.ErrorResponse
// @Failure      404                   {object}  response.ErrorResponse
// @Router       /api/auth/send-code [post]
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

// @Summary      Получить информацию о текущем пользователе
// @Description  Возвращает данные пользователя по JWT из заголовка
// @Tags         Auth
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  dto.UserResponse
// @Failure      401  {object}  response.ErrorResponse
// @Router       /api/users/me [get]
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

func (h *Handler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	var req dto.GoogleLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, ErrInvalidRequest.Error())
		return
	}

	ctx := r.Context()
	accessToken, refreshToken, usr, err := h.service.LoginWithGoogle(ctx, req.IDToken)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
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
