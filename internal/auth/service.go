package auth

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"task-planner/internal/email"
	"task-planner/internal/user"
	"task-planner/pkg/security"
	"time"
)

type Service struct {
	userService user.Service
	emailSvc    email.EmailService
	emailRepo   email.EmailRepository
	tokenRepo   TokenRepository
	JWTConfig   JWTConfig
}

func NewService(u user.Service, e email.EmailService, er email.EmailRepository, tr TokenRepository, jwtCft JWTConfig) *Service {
	return &Service{
		userService: u,
		emailSvc:    e,
		emailRepo:   er,
		tokenRepo:   tr,
		JWTConfig:   jwtCft,
	}
}

func (s *Service) RegisterEmail(ctx context.Context, email, password, name string) error {
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

	s.emailSvc.SendVerificationEmailAsync(email, code)

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

func (s *Service) VerifyEmail(ctx context.Context, email, code string) error {
	user, err := s.userService.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}

	if user == nil {
		return fmt.Errorf("user not found")
	}

	verificationCode, err := s.emailRepo.GetVerificationCode(ctx, user.ID)
	if err != nil {
		return err
	}

	if verificationCode == nil {
		return fmt.Errorf("Verification code not found")
	}

	if time.Now().After(verificationCode.ExpiresAt) {
		return fmt.Errorf("Verification code expired")
	}

	if verificationCode.Code != code {
		return fmt.Errorf("Invalid verification code")
	}

	err = s.userService.MarkEmailAsVerified(ctx, user.ID)
	if err != nil {
		return err
	}

	err = s.emailRepo.DeleteVerificationCode(ctx, user.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error) {
	usr, err := s.userService.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", err
	}
	if usr == nil {
		return "", "", ErrInvalidCredentials
	}

	if err := security.ComparePasswords(usr.PasswordHash, password); err != nil {
		return "", "", ErrInvalidCredentials
	}

	accessToken, refreshToken, err = GenerateTokenPair(usr.ID, usr.Email, s.JWTConfig)
	if err != nil {
		return "", "", fmt.Errorf("failed to save refresh token: %w", err)
	}
	return accessToken, refreshToken, nil
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
	if err := s.tokenRepo.DeleteRefreshToken(ctx, refreshToken); err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}
	return nil
}
