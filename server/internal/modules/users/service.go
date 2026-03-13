package users

import (
	"context"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}
func (service *Service) Create(ctx context.Context, user *User) error {
	return service.repo.Create(ctx, user)
}

func (service *Service) GetUser(ctx context.Context, id uuid.UUID) (*User, error) {
	return service.repo.GetByID(ctx, id)
}
