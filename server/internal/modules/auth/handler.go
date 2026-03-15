package auth

import (
	"errors"
	"net/http"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/auth"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/apierror"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/httputil"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	app     *app.Application
	service *Service
}

func NewHandler(app *app.Application, service *Service) *Handler {
	return &Handler{
		app:     app,
		service: service,
	}
}

// registerUserHandler godoc
//
//	@Summary		Register a user
//	@Description	Registers a new user and creates a verification token
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		RegisterUserPayload	true	"User credentials"
//	@Success		201		{object}	UserWithToken		"User registered"
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Router			/auth/register [post]
func (handler *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	logger := handler.app.Logger

	// Parse the request body & extract  the payload
	var payload RegisterUserPayload

	if err := httputil.ReadJSON(w, r, &payload); err != nil {
		apierror.BadRequest(logger, w, r, err)
		return
	}

	ctx := r.Context()
	userToken, err := handler.service.RegisterUser(ctx, payload)
	if err != nil {
		switch err {
		case apierror.ErrBadRequest:
			apierror.BadRequest(logger, w, r, err)
			return
		case apierror.ErrNotFound:
			apierror.NotFound(logger, w, r, err)
			return
		default:
			apierror.InternalServerError(logger, w, r, err)
			return
		}
	}

	userTokenResp := &UserWithTokenResponse{
		User:  userToken.User.ToResponse(),
		Token: userToken.Token,
	}

	// Attach the data to the response body
	if err := httputil.JSONResponse(w, http.StatusCreated, userTokenResp, "User has Registered successfully, This is his data."); err != nil {
		apierror.InternalServerError(logger, w, r, err)
		return
	}
}

// activateUserHandler godoc
//
//	@Summary		Activate a user account
//	@Description	Activates a user using a verification token
//	@Tags			authentication
//	@Produce		json
//	@Param			token	path		string	true	"Activation token"
//	@Success		204		{string}	string	"User activated"
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Router			/auth/activate/{token} [post]
func (handler *Handler) ActivateUser(w http.ResponseWriter, r *http.Request) {
	logger := handler.app.Logger

	// Extract the token From URL query
	token := chi.URLParam(r, "token")

	if token == "" {
		apierror.BadRequest(logger, w, r, errors.New("Verification Token is Required to activate your account."))
		return
	}

	ctx := r.Context()
	err := handler.service.ActivateUser(ctx, token)

	if err != nil {
		switch err {
		case apierror.ErrNotFound:
			apierror.NotFound(logger, w, r, err)
			return
		default:
			apierror.InternalServerError(logger, w, r, err)
			return
		}
	}

	// send a response
	if err := httputil.JSONResponse(w, http.StatusNoContent, nil, "User was activated successfully."); err != nil {
		apierror.InternalServerError(logger, w, r, err)
		return
	}
}

// Authenticate godoc
//
//	@Summary		Authenticate user
//	@Description	Authenticates a user using email and password, returns the user data and sets an HttpOnly JWT cookie.
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		AuthenticatePayload	true	"Login credentials"
//	@Success		200		{object}	UserWithToken		"User authenticated successfully"
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Router			/auth/login [post]
func (handler *Handler) AuthenticateUser(w http.ResponseWriter, r *http.Request) {
	logger := handler.app.Logger
	var payload AuthenticatePayload

	// Read the request body
	if err := httputil.ReadJSON(w, r, &payload); err != nil {
		apierror.BadRequest(logger, w, r, err)
		return
	}

	ctx := r.Context()

	// Call the authenticate service method
	userToken, err := handler.service.Authenticate(ctx, payload)

	if err != nil {
		switch err {
		case apierror.ErrBadRequest:
			apierror.BadRequest(logger, w, r, err)
			return
		case apierror.ErrNotFound:
			apierror.NotFound(logger, w, r, err)
			return
		case apierror.ErrUnauthorized:
			apierror.Unauthorized(logger, w, r, err)
		default:
			apierror.InternalServerError(logger, w, r, err)
			return
		}
	}
	// Set a cookie with the JWT (90 days)
	auth.SetCookie(w, auth.AuthTokenCookieKey, userToken.Token, auth.AuthTokenCookieExp)

	userTokenResp := &UserWithTokenResponse{
		User:  userToken.User.ToResponse(),
		Token: userToken.Token,
	}

	// send a response
	if err := httputil.JSONResponse(w, http.StatusOK, userTokenResp, "User is authenticated successfully."); err != nil {
		apierror.InternalServerError(logger, w, r, err)
		return
	}
}
