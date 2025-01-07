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
}

func NewService(userService user.UserService, emailSvc email.EmailService) *Service {
	return &Service{
		userService: userService,
		emailSvc:    emailSvc,
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

	if err := s.userService.CreateUser(email, hashedPassword, name); err != nil {
		return err
	}
	if err := s.emailSvc.SendVerificationEmail(email, "123"); err != nil {
		return err
	}

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
