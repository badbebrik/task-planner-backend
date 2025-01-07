package user

type UserService interface {
	CreateUser(email, passwordHash, name string) error
	UserExists(email string) (bool, error)
}

type Service struct {
	repo UserRepository
}

func NewService(repo UserRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateUser(email, passwordHash, name string) error {
	exists, err := s.UserExists(email)
	if err != nil {
		return err
	}
	if exists {
		return ErrUserAlreadyExists
	}

	user := &User{
		Email:        email,
		PasswordHash: passwordHash,
		Name:         name,
	}
	return s.repo.CreateUser(user)
}

func (s *Service) UserExists(email string) (bool, error) {
	return s.repo.UserExists(email)
}
