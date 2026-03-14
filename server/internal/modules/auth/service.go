package auth

import (
	"context"
	"time"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/auth"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/config"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/users"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/validator"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/apierror"
	"github.com/golang-jwt/jwt/v5"
)

type Service struct {
	userService   *users.Service
	authenticator *auth.JWTAuthenticator
	cfg           *config.Config
}

func NewService(userService *users.Service, authenticator *auth.JWTAuthenticator, cfg config.Config) *Service {
	return &Service{
		userService:   userService,
		authenticator: authenticator,
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
		return nil, err
	}

	// create user
	createUserPayload := &users.CreateUserPayload{
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
		Email:     payload.Email,
		Password:  payload.Password,
	}
	user, err := service.userService.Create(ctx, createUserPayload)
	if err != nil {
		return nil, err
	}

	// generate JWT Token

	claims := UserClaims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:   service.cfg.Authenticator.JWT.Iss,
			Audience: []string{service.cfg.Authenticator.JWT.Aud},
			ExpiresAt: jwt.NewNumericDate(
				time.Now().Add(service.cfg.Authenticator.JWT.MailTokenExp),
			),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}
	token, err := service.authenticator.GenerateToken(claims)
	if err != nil {
		return nil, err
	}
	// form User with context
	userToken := &UserWithToken{
		User:  user,
		Token: token,
	}
	// send email

	return userToken, nil
}
