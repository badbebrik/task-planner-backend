package auth

import (
	"errors"
	"task-planner/internal/email"
	"task-planner/internal/user"
	"task-planner/pkg/security"
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
