package auth

import (
	"net/http"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/apierror"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/httputil"
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

	// Attach the data to the response body
	if err := httputil.JSONResponse(w, http.StatusCreated, userToken, "User has Registered successfully, This is his data."); err != nil {
		apierror.InternalServerError(logger, w, r, err)
		return
	}
}

func (handler *Handler) ActivateUser(w http.ResponseWriter, r *http.Request) {
	// Extract the token
	
	// Extract the claims 



}

func (handler *Handler) Authenticate(w http.ResponseWriter, r *http.Request) {

}
