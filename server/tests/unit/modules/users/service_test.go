package users_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/users"
	apierror "github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mocks
type mockUsersRepo struct {
	mock.Mock
}

func (m *mockUsersRepo) Create(ctx context.Context, user *users.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUsersRepo) GetByID(ctx context.Context, id uuid.UUID) (*users.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*users.User), args.Error(1)
}

func (m *mockUsersRepo) GetByEmail(ctx context.Context, email string) (*users.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*users.User), args.Error(1)
}

func (m *mockUsersRepo) ActivateUser(ctx context.Context, user *users.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUsersRepo) GetFromToken(ctx context.Context, token []byte) (*users.User, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*users.User), args.Error(1)
}

func (m *mockUsersRepo) WithTx(tx *sql.Tx) users.IRepository {
	return m
}

func TestUsersService_Create(t *testing.T) {
	mockRepo := new(mockUsersRepo)
	service := users.NewService(mockRepo)

	t.Run("successful user creation", func(t *testing.T) {
		payload := &users.CreateUserPayload{
			FirstName: "Jane",
			LastName:  "Doe",
			Email:     "jane@example.com",
			Password:  "securepassword",
		}

		mockRepo.On("GetByEmail", mock.Anything, payload.Email).Return(nil, apierror.ErrNotFound)
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*users.User")).Return(nil)

		user, err := service.Create(context.Background(), payload)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, payload.Email, user.Email)
		mockRepo.AssertExpectations(t)
	})

	t.Run("user creation failed - email exists", func(t *testing.T) {
		payload := &users.CreateUserPayload{
			Email: "existing@example.com",
		}

		mockRepo.On("GetByEmail", mock.Anything, payload.Email).Return(&users.User{}, nil)

		user, err := service.Create(context.Background(), payload)

		assert.Error(t, err)
		assert.Equal(t, apierror.ErrConflict, err)
		assert.Nil(t, user)
	})
}

func TestUsersService_GetUser(t *testing.T) {
	mockRepo := new(mockUsersRepo)
	service := users.NewService(mockRepo)

	t.Run("successful get user by ID", func(t *testing.T) {
		id := uuid.New()
		expectedUser := &users.User{ID: id.String(), Email: "test@example.com"}

		mockRepo.On("GetByID", mock.Anything, id).Return(expectedUser, nil)

		user, err := service.GetUser(context.Background(), id.String())

		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)
	})

	t.Run("failed get user - invalid UUID", func(t *testing.T) {
		user, err := service.GetUser(context.Background(), "invalid-uuid")

		assert.Error(t, err)
		assert.Nil(t, user)
	})
}
