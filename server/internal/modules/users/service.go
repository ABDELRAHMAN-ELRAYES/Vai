package users

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetUser(id string) (*User, error) {
	return s.repo.FindByID(id)
}