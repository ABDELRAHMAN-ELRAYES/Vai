package users

import (
	"context"
	"errors"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/validator"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/apierror"
	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}
func (service *Service) Create(ctx context.Context, payload *CreateUserPayload) (*User, error) {
	// Validate request body
	if err := validator.Validate.Struct(payload); err != nil {
		return nil, apierror.ErrBadRequest
	}
	// Check if the email already exists
	_, err := service.repo.GetByEmail(ctx, payload.Email)
	if err != nil {
		if !errors.Is(err, apierror.ErrNotFound) {
			return nil, err
		}
	} else {
		return nil, apierror.ErrConflict
	}

	// Create user model
	user := &User{
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
		Email:     payload.Email,
	}
	// - Hash the user plaintext password
	if err := user.Password.Set(payload.Password); err != nil {
		return nil, err
	}

	//  Store user in the DB
	if err := service.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil

}

func (service *Service) GetUser(ctx context.Context, id string) (*User, error) {
	uID, err := uuid.Parse(id)
	if err != nil {

		return nil, apierror.ErrBadRequest
	}
	return service.repo.GetByID(ctx, uID)
}
func (service *Service) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return service.repo.GetByEmail(ctx, email)
}
