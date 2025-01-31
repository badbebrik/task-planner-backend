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

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
)

type Service struct {
	userService user.Service
	emailSvc    email.EmailService
	emailRepo   email.EmailRepository
}

func NewService(u user.Service, e email.EmailService, er email.EmailRepository) *Service {
	return &Service{
		userService: u,
		emailSvc:    e,
		emailRepo:   er,
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
