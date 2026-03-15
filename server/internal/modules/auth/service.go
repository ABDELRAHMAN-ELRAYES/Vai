package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/auth"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/config"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/db"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/users"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/validator"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/apierror"
	"github.com/golang-jwt/jwt/v5"
)

type Service struct {
	db            *sql.DB
	repo          *Repository
	userService   *users.Service
	authenticator *auth.JWTAuthenticator
	cfg           *config.Config
}

func NewService(
	db *sql.DB,
	repo *Repository,
	userService *users.Service,
	authenticator *auth.JWTAuthenticator,
	cfg *config.Config,
) *Service {
	return &Service{
		db:            db,
		repo:          repo,
		userService:   userService,
		authenticator: authenticator,
		cfg:           cfg,
	}
}

func (service *Service) RegisterUser(ctx context.Context, payload RegisterUserPayload) (*UserWithToken, error) {
	// Validate the payload
	if err := validator.Validate.Struct(payload); err != nil {
		return nil, apierror.ErrBadRequest
	}
	// check if the user exists by email
	_, err := service.userService.GetUserByEmail(ctx, payload.Email)

	if err != nil {
		if !errors.Is(err, apierror.ErrNotFound) {
			return nil, err
		}
	}

	// create user
	user := &users.User{
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
		Email:     payload.Email,
	}

	if err = user.Password.Set(payload.Password); err != nil {
		return nil, err
	}
	token := auth.GenerateRandomToken()
	hashedToken := auth.HashToken(token)

	// Create a new user + token (with tranasaction)
	err = db.WithTx(service.db, ctx, func(tx *sql.Tx) error {
		userRepo := users.NewRepository(tx)
		authRepo := NewRepository(tx)

		err = userRepo.Create(ctx, user)
		if err != nil {
			return err
		}
		token := &Token{
			UserID:    user.ID,
			Token:     hashedToken,
			ExpiredAt: time.Now().Add(service.cfg.Authenticator.JWT.MailTokenExp),
		}

		err = authRepo.CreateToken(ctx, token)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	// form User with context
	userToken := &UserWithToken{
		User:  user,
		Token: token,
	}
	//TODO: send email

	return userToken, nil
}

func (service *Service) ActivateUser(ctx context.Context, token string) error {
	hashedToken := auth.HashToken(token)

	return db.WithTx(service.db, ctx, func(tx *sql.Tx) error {
		userRepo := users.NewRepository(tx)
		authRepo := NewRepository(tx)

		// Get  User by The sent token
		user, err := userRepo.GetFromToken(ctx, hashedToken)
		if err != nil {
			return err
		}

		// Activate user account
		user.IsActive = true
		if err := userRepo.ActivateUser(ctx, user); err != nil {
			return err
		}

		// Clean up the token
		if err := authRepo.CleanUpToken(ctx, user.ID); err != nil {
			return err
		}

		return nil
	})
}

func (service *Service) Authenticate(ctx context.Context, payload AuthenticatePayload) (*UserWithToken, error) {

	// Validate the payload
	if err := validator.Validate.Struct(payload); err != nil {
		return nil, apierror.ErrBadRequest
	}

	// Check if the user exists
	user, err := service.userService.GetUserByEmail(ctx, payload.Email)
	if err != nil {
		return nil, err
	}
	// Compare the password
	if err := user.Password.Compare(payload.Password); err != nil {
		return nil, apierror.ErrUnauthorized
	}
	// Formm the JWT claims
	claims := UserClaims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(service.cfg.Authenticator.JWT.SessionExp)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    service.cfg.Authenticator.JWT.Iss,
		},
	}

	// Generate a JWT
	token, err := service.authenticator.GenerateToken(claims)

	if err != nil {
		return nil, err
	}

	userWithToken := &UserWithToken{
		User:  user,
		Token: token,
	}

	return userWithToken, nil
}
