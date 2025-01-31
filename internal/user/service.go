package user

import "context"

type Service interface {
	CreateUser(ctx context.Context, email, passwordHash, name string) (int64, error)
	UserExists(ctx context.Context, email string) (bool, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	MarkEmailAsVerified(ctx context.Context, id int64) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateUser(ctx context.Context, email, passwordHash, name string) (int64, error) {
	exists, err := s.repo.UserExists(ctx, email)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, ErrUserAlreadyExists
	}

	user := &User{
		Email:        email,
		PasswordHash: passwordHash,
		Name:         name,
	}

	userID, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (s *service) UserExists(ctx context.Context, email string) (bool, error) {
	return s.repo.UserExists(ctx, email)
}

func (s *service) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return s.repo.GetUserByEmail(ctx, email)
}

func (s *service) MarkEmailAsVerified(ctx context.Context, id int64) error {
	return s.repo.MarkEmailAsVerified(ctx, id)
}
