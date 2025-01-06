package user

import (
	"errors"
	"task-planner/pkg/security"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateUser(email, password, name string) error {
	exists, err := s.repo.UserExists(email)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("user already exists")
	}
	hashedPassword, err := security.HashPassword(password)
	if err != nil {
		return err
	}
	user := &User{
		Email:           email,
		PasswordHash:    hashedPassword,
		Name:            name,
		IsEmailVerified: false,
	}
	if err := s.repo.CreateUser(user); err != nil {
		return err
	}
	return nil
}
