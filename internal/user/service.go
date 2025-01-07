package user

type UserService interface {
	CreateUser(email, passwordHash, name string) (int, error)
	UserExists(email string) (bool, error)
}

type Service struct {
	repo UserRepository
}

func NewService(repo UserRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateUser(email, passwordHash, name string) (int, error) {
	exists, err := s.UserExists(email)
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

	userID, err := s.repo.CreateUser(user)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (s *Service) UserExists(email string) (bool, error) {
	return s.repo.UserExists(email)
}
