package auth

import (
	"errors"
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
