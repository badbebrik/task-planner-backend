package auth

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"task-planner/internal/email"
	"task-planner/internal/user"
	"task-planner/pkg/config"
	"task-planner/pkg/security"
	"time"
)

type Service struct {
	userService  user.Service
	emailService email.EmailService
	emailRepo    email.EmailRepository
	tokenRepo    TokenRepository
	JWTConfig    config.JWTConfig
}

func NewService(u user.Service, e email.EmailService, er email.EmailRepository, tr TokenRepository, jwtCft config.JWTConfig) *Service {
	return &Service{
		userService:  u,
		emailService: e,
		emailRepo:    er,
		tokenRepo:    tr,
		JWTConfig:    jwtCft,
	}
}

func (s *Service) Signup(ctx context.Context, email, password, name string) error {
	exists, err := s.userService.UserExists(ctx, email)
	if err != nil {
		return err
	}

	if exists {
		return ErrUserAlreadyExists
	}

	hashedPassword, err := security.HashPassword(password)
	if err != nil {
		return err
	}

	userID, err := s.userService.CreateUser(ctx, email, hashedPassword, name)
	if err != nil {
		return err
	}

	code := GenerateVerificationCode()
	expiresAt := time.Now().Add(10 * time.Minute)

	err = s.emailRepo.SaveVerificationCode(ctx, userID, code, expiresAt)
	if err != nil {
		return err
	}

	s.emailService.SendVerificationEmailAsync(email, code)

	return nil
}

func GenerateVerificationCode() string {
	rand.NewSource(time.Now().UnixNano())
	letters := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	code := make([]byte, 4)
	for i := range code {
		code[i] = letters[rand.Intn(len(letters))]
	}
	return string(code)
}

func (s *Service) VerifyEmailAndGetTokens(ctx context.Context, email, code string) (string, string, error) {
	usr, err := s.userService.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", err
	}
	if usr == nil {
		return "", "", fmt.Errorf("user not found")
	}

	verificationCode, err := s.emailRepo.GetVerificationCode(ctx, usr.ID)
	if err != nil {
		return "", "", err
	}
	if verificationCode == nil {
		return "", "", fmt.Errorf("verification code not found")
	}
	if time.Now().After(verificationCode.ExpiresAt) {
		return "", "", fmt.Errorf("verification code expired")
	}
	if verificationCode.Code != code {
		return "", "", fmt.Errorf("invalid verification code")
	}

	err = s.userService.MarkEmailAsVerified(ctx, usr.ID)
	if err != nil {
		return "", "", err
	}
	_ = s.emailRepo.DeleteVerificationCode(ctx, usr.ID)

	accessToken, refreshToken, err := GenerateTokenPair(usr.ID, usr.Email, s.JWTConfig)
	if err != nil {
		return "", "", err
	}
	refreshExp := time.Now().Add(s.JWTConfig.RefreshTTL)
	err = s.tokenRepo.SaveRefreshToken(ctx, usr.ID, refreshToken, refreshExp)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (string, string, *user.User, error) {
	usr, err := s.userService.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", nil, err
	}
	if usr == nil {
		return "", "", nil, ErrInvalidCredentials
	}
	if err := security.ComparePasswords(usr.PasswordHash, password); err != nil {
		return "", "", nil, ErrInvalidCredentials
	}
	accessToken, refreshToken, err := GenerateTokenPair(usr.ID, usr.Email, s.JWTConfig)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate tokens: %w", err)
	}
	refreshExp := time.Now().Add(s.JWTConfig.RefreshTTL)
	if err := s.tokenRepo.SaveRefreshToken(ctx, usr.ID, refreshToken, refreshExp); err != nil {
		return "", "", nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return accessToken, refreshToken, usr, nil
}

func (s *Service) RefreshTokens(ctx context.Context, refreshToken string) (newAccess, newRefresh string, err error) {
	claims, err := ValidateToken(refreshToken, s.JWTConfig.RefreshSecret)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token: %w", err)
	}

	existingToken, err := s.tokenRepo.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", "", err
	}
	if existingToken == nil {
		return "", "", errors.New("refresh token not found or already revoked")
	}
	if time.Now().After(existingToken.ExpiresAt) {
		_ = s.tokenRepo.DeleteRefreshToken(ctx, refreshToken)
		return "", "", errors.New("refresh token expired")
	}

	newAccess, newRefresh, err = GenerateTokenPair(claims.UserID, claims.Email, s.JWTConfig)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate new token pair: %w", err)
	}
	if err := s.tokenRepo.DeleteRefreshToken(ctx, refreshToken); err != nil {
		return "", "", fmt.Errorf("failed to delete old refresh token: %w", err)
	}
	refreshExp := time.Now().Add(s.JWTConfig.RefreshTTL)
	if err := s.tokenRepo.SaveRefreshToken(ctx, claims.UserID, newRefresh, refreshExp); err != nil {
		return "", "", fmt.Errorf("failed to save new refresh token: %w", err)
	}

	return newAccess, newRefresh, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	return s.tokenRepo.DeleteRefreshToken(ctx, refreshToken)
}

func (s *Service) SendVerificationCode(ctx context.Context, email string) error {
	usr, err := s.userService.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	if usr == nil {
		return ErrUserNotFound
	}

	if usr.IsEmailVerified {
		return errors.New("email already verified")
	}

	code := GenerateVerificationCode()
	expiresAt := time.Now().Add(10 * time.Minute)

	err = s.emailRepo.SaveVerificationCode(ctx, usr.ID, code, expiresAt)
	if err != nil {
		return err
	}

	s.emailService.SendVerificationEmailAsync(email, code)
	return nil
}
