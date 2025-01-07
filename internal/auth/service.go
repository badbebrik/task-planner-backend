package auth

import (
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
)

type Service struct {
	userService user.UserService
	emailSvc    email.EmailService
	emailRepo   email.EmailRepository
}

func NewService(userService user.UserService, emailSvc email.EmailService, emailRepo email.EmailRepository) *Service {
	return &Service{
		userService: userService,
		emailSvc:    emailSvc,
		emailRepo:   emailRepo,
	}
}

func (s *Service) RegisterEmail(email, password, name string) error {
	exists, err := s.userService.UserExists(email)
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

	userID, err := s.userService.CreateUser(email, hashedPassword, name)
	if err != nil {
		return err
	}

	code := GenerateVerificationCode()
	expiresAt := time.Now().Add(10 * time.Minute)

	err = s.emailRepo.SaveVerificationCode(userID, code, expiresAt)
	if err != nil {
		return err
	}

	s.emailSvc.SendVerificationEmailAsync(email, code)

	return nil
}

func GenerateVerificationCode() string {
	rand.NewSource(time.Now().UnixNano())
	letters := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	code := make([]byte, 6)
	for i := range code {
		code[i] = letters[rand.Intn(len(letters))]
	}
	return string(code)
}

func (s *Service) VerifyEmail(email, code string) error {
	user, err := s.userService.GetUserByEmail(email)
	if err != nil {
		return err
	}

	if user == nil {
		return fmt.Errorf("user not found")
	}

	verificationCode, err := s.emailRepo.GetVerificationCode(user.ID)
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

	err = s.userService.MarkEmailAsVerified(user.ID)
	if err != nil {
		return err
	}

	err = s.emailRepo.DeleteVerificationCode(user.ID)
	if err != nil {
		return err
	}

	return nil
}
