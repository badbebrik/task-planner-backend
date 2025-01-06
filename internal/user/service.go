package user

import (
	"errors"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateUser(email, passwordHash, name string) error {
	exists, err := s.repo.UserExists(email)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("user already exists")
	}

	user := &User{
		Email:           email,
		PasswordHash:    passwordHash,
		Name:            name,
		IsEmailVerified: false,
	}
	if err := s.repo.CreateUser(user); err != nil {
		return err
	}
	return nil
}
