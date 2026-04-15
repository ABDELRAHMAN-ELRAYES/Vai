package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	authInternal "github.com/ABDELRAHMAN-ELRAYES/Vai/internal/auth"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/config"
	dbPkg "github.com/ABDELRAHMAN-ELRAYES/Vai/internal/db"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/mailer"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/users"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/validator"
	apierror "github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/errors"
	"github.com/golang-jwt/jwt/v5"
)



type Service struct {
	database      *sql.DB
	repo          RepositoryInterface
	userService   users.ServiceInterface
	authenticator *authInternal.JWTAuthenticator
	cfg           *config.Config
	mailer        mailer.Client
}

func NewService(
	db *sql.DB,
	repo RepositoryInterface,
	userService users.ServiceInterface,
	authenticator *authInternal.JWTAuthenticator,
	cfg *config.Config,
	mailer mailer.Client,
) *Service {
	return &Service{
		database:      db,
		repo:          repo,
		userService:   userService,
		authenticator: authenticator,
		cfg:           cfg,
		mailer:        mailer,
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
		if !errors.Is(err, apierror.ErrNotFound) && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	} else {
		// User found, conflict
		return nil, apierror.ErrConflict
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
	token := authInternal.GenerateRandomToken()
	hashedToken := authInternal.HashToken(token)

	// Create a new user + token (with tranasaction)
	err = dbPkg.WithTx(service.database, ctx, func(tx *sql.Tx) error {
		userServ := service.userService.WithTx(tx)
		authRepo := service.repo.WithTx(tx)

		err = userServ.CreateWithModel(ctx, user)
		if err != nil {
			return err
		}
		
		tokenModel := &Token{
			UserID:    user.ID,
			Token:     hashedToken,
			ExpiredAt: time.Now().Add(service.cfg.Authenticator.JWT.MailTokenExp),
		}

		err = authRepo.CreateToken(ctx, tokenModel)
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
	//send email
	activationURL := fmt.Sprintf("%s/confirm/%s", service.cfg.FrontendURL, token)

	mailTmplData := &mailer.VerifyEmailData{
		Name:      fmt.Sprintf("%s %s", user.FirstName, user.LastName),
		VerifyURL: activationURL,
		SentAt:    time.Now(),
	}

	res, err := service.mailer.Send(ctx, "verify_email", user.Email, mailTmplData)
	if err != nil {
		return nil, err
	}
	if res != mailer.SMTPSuccessCode {
		return nil, apierror.ErrMailSendFailed
	}

	return userToken, nil
}

func (service *Service) ActivateUser(ctx context.Context, token string) error {
	hashedToken := authInternal.HashToken(token)

	return dbPkg.WithTx(service.database, ctx, func(tx *sql.Tx) error {
		userServ := service.userService.WithTx(tx)
		authRepo := service.repo.WithTx(tx)

		// Get  User by The sent token
		user, err := userServ.GetFromToken(ctx, hashedToken)
		if err != nil {
			return err
		}

		// Activate user account
		user.IsActive = true
		if err := userServ.ActivateUser(ctx, user); err != nil {
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
			Audience:  []string{service.cfg.Authenticator.JWT.Aud},
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
