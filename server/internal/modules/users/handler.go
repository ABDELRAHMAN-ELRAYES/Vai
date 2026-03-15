package users

import (
	"net/http"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"
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

// createUserHandler godoc
//
//	@Summary		Create a user
//	@Description	Creates a new user account
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		CreateUserPayload	true	"User data"
//	@Success		201		{object}	User
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Router			/users [post]
func (handler *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	logger := handler.app.Logger

	// Parse the reqquest body
	var payload CreateUserPayload

	if err := httputil.ReadJSON(w, r, &payload); err != nil {
		apierror.BadRequest(logger, w, r, err)
		return
	}
	// Extract request context
	ctx := r.Context()

	// Create User
	user, err := handler.service.Create(ctx, &payload)
	if err != nil {
		switch err {
		case err, apierror.ErrBadRequest:
			apierror.BadRequest(logger, w, r, err)
			return
		case apierror.ErrConflict:
			apierror.Conflict(logger, w, r, err)
			return
		default:
			apierror.InternalServerError(logger, w, r, err)
			return
		}
	}

	userResp := user.ToResponse()

	// Attach the data to the response body
	if err := httputil.JSONResponse(w, http.StatusCreated, userResp, "User has been created successfully, This is his data."); err != nil {
		apierror.InternalServerError(logger, w, r, err)
		return
	}

}

// GetUser godoc
//
//	@Summary		Fetches a user profile
//	@Description	Fetches a user profile by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		string	true	"User ID"
//	@Success		200		{object}	User
//	@Failure		400		{object}	error
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/{userID} [get]
func (handler *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	logger := handler.app.Logger

	id := chi.URLParam(r, "userID")

	ctx := r.Context()

	user, err := handler.service.GetUser(ctx, id)

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
	userResp := user.ToResponse()

	// Attach the data to the response body
	if err := httputil.JSONResponse(w, http.StatusOK, userResp, "This is the Data of the user with entered ID"); err != nil {
		apierror.InternalServerError(logger, w, r, err)
		return
	}
}
