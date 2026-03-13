package users

import (
	"net/http"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/validator"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/apierror"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/httputil"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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
	// Parse the reqquest body
	var payload CreateUserPayload

	if err := httputil.ReadJSON(w, r, &payload); err != nil {
		apierror.BadRequest(handler.app.Logger, w, r, err)
		return
	}

	// Validate request body
	if err := validator.Validate.Struct(payload); err != nil {
		apierror.BadRequest(handler.app.Logger, w, r, err)
		return
	}

	// Create user model
	user := &User{
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
		Email:     payload.Email,
	}
	// - Hash the user plaintext password
	if err := user.Password.Set(payload.Password); err != nil {
		apierror.InternalServerError(handler.app.Logger, w, r, err)
		return
	}
	// - Extract request context
	ctx := r.Context()

	//  Store user in the DB
	if err := handler.service.Create(ctx, user); err != nil {
		apierror.InternalServerError(handler.app.Logger, w, r, err)
		return
	}

	// Attach the data to the response body

	if err := httputil.JSONResponse(w, http.StatusCreated, user, "User has been created successfully, This is his data."); err != nil {
		apierror.InternalServerError(handler.app.Logger, w, r, err)
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

	id := chi.URLParam(r, "userID")
	uID, err := uuid.Parse(id)
	if err != nil {
		apierror.BadRequest(handler.app.Logger, w, r, err)
		return
	}

	ctx := r.Context()

	user, err := handler.service.GetUser(ctx, uID)

	if err != nil {
		switch err {
		case apierror.ErrNotFound:
			apierror.NotFound(handler.app.Logger, w, r, err)
		default:
			apierror.InternalServerError(handler.app.Logger, w, r, err)

		}
		return
	}

	// Attach the data to the response body
	if err := httputil.JSONResponse(w, http.StatusOK, user, "This is the Data of the user with entered ID"); err != nil {
		apierror.InternalServerError(handler.app.Logger, w, r, err)
		return
	}
}
