package auth_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/auth"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/config"
	authModule "github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/auth"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/users"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mocks
type mockAuthRepo struct {
	mock.Mock
}

func (m *mockAuthRepo) CreateToken(ctx context.Context, token *authModule.Token) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *mockAuthRepo) CleanUpToken(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *mockAuthRepo) WithTx(tx *sql.Tx) authModule.IRepository {
	m.Called(tx)
	return m
}

type mockUsersService struct {
	mock.Mock
}

func (m *mockUsersService) GetUserByEmail(ctx context.Context, email string) (*users.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*users.User), args.Error(1)
}

func (m *mockUsersService) GetFromToken(ctx context.Context, token []byte) (*users.User, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*users.User), args.Error(1)
}

func (m *mockUsersService) ActivateUser(ctx context.Context, user *users.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUsersService) CreateWithModel(ctx context.Context, user *users.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUsersService) WithTx(tx *sql.Tx) users.IService {
	m.Called(tx)
	return m
}

type mockMailer struct {
	mock.Mock
}

func (m *mockMailer) Send(ctx context.Context, templateFile, toEmail string, data any) (int, error) {
	args := m.Called(ctx, templateFile, toEmail, data)
	return args.Int(0), args.Error(1)
}

func TestAuthService_Authenticate(t *testing.T) {
	db, _, _ := sqlmock.NewWithDSN("auth_authenticate_db")
	mockRepo := new(mockAuthRepo)
	mockUsers := new(mockUsersService)
	mockMail := new(mockMailer)

	cfg := &config.Config{
		Authenticator: config.AuthenticatorConfig{
			JWT: config.JWTConfig{
				Secret:     "secret",
				Iss:        "iss",
				Aud:        "aud",
				SessionExp: time.Hour,
			},
		},
	}
	authenticator := auth.NewJWTuthenticator(cfg.Authenticator.JWT.Secret, cfg.Authenticator.JWT.Iss, cfg.Authenticator.JWT.Aud)

	service := authModule.NewService(db, mockRepo, mockUsers, authenticator, cfg, mockMail)

	t.Run("successful authentication", func(t *testing.T) {
		payload := authModule.AuthenticatePayload{
			Email:    "test@example.com",
			Password: "password123",
		}

		user := &users.User{
			Email: "test@example.com",
		}
		user.Password.Set("password123")

		mockUsers.On("GetUserByEmail", mock.Anything, payload.Email).Return(user, nil)

		res, err := service.Authenticate(context.Background(), payload)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, user.Email, res.User.Email)
		assert.NotEmpty(t, res.Token)
		mockUsers.AssertExpectations(t)
	})

	t.Run("failed authentication - wrong password", func(t *testing.T) {
		payload := authModule.AuthenticatePayload{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}

		user := &users.User{
			Email: "test@example.com",
		}
		user.Password.Set("password123")

		mockUsers.On("GetUserByEmail", mock.Anything, payload.Email).Return(user, nil)

		res, err := service.Authenticate(context.Background(), payload)

		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestAuthService_RegisterUser(t *testing.T) {
	t.Run("successful registration", func(t *testing.T) {
		db, mockDB, _ := sqlmock.New()
		mockRepo := new(mockAuthRepo)
		mockUsers := new(mockUsersService)
		mockMail := new(mockMailer)

		cfg := &config.Config{
			FrontendURL: "http://localhost:3000",
			Authenticator: config.AuthenticatorConfig{
				JWT: config.JWTConfig{
					MailTokenExp: time.Hour,
				},
			},
		}

		service := authModule.NewService(db, mockRepo, mockUsers, nil, cfg, mockMail)

		payload := authModule.RegisterUserPayload{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
			Password:  "password123",
		}

		mockUsers.On("GetUserByEmail", mock.Anything, payload.Email).Return(nil, sql.ErrNoRows)

		// Transaction expectations
		mockDB.ExpectBegin()
		mockUsers.On("WithTx", mock.Anything).Return(mockUsers)
		mockRepo.On("WithTx", mock.Anything).Return(mockRepo)

		mockUsers.On("CreateWithModel", mock.Anything, mock.AnythingOfType("*users.User")).Return(nil)
		mockRepo.On("CreateToken", mock.Anything, mock.AnythingOfType("*auth.Token")).Return(nil)
		mockDB.ExpectCommit()

		mockMail.On("Send", mock.Anything, "verify_email", payload.Email, mock.Anything).Return(250, nil)

		res, err := service.RegisterUser(context.Background(), payload)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, payload.Email, res.User.Email)
		assert.NotEmpty(t, res.Token)

		mockUsers.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
		mockMail.AssertExpectations(t)
		assert.NoError(t, mockDB.ExpectationsWereMet())
	})
}
