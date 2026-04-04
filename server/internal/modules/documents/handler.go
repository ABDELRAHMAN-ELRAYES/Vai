package documents

import (
	"net/http"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/shared"
	sharedDocs "github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/shared/modules/documents"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/users"
	apierror "github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/errors"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/httputil"
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

func (handler *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	// Get uploaded file from context
	result, ok := sharedDocs.GetUploadedFile(r)
	if !ok {
		apierror.InternalServerError(handler.app.Logger, w, r, apierror.ErrNoFileProvided)
		return
	}

	// Get user data from context
	user, ok := r.Context().Value(shared.UserCtxKey).(*users.User)
	if !ok {
		apierror.Unauthorized(handler.app.Logger, w, r, apierror.ErrUnauthorized)
		return
	}

	ownerID, err := uuid.Parse(user.ID)
	if err != nil {
		apierror.InternalServerError(handler.app.Logger, w, r, err)
		return
	}

	// Save document metadata
	doc, err := handler.service.CreateDocument(r.Context(), ownerID, result)
	if err != nil {
		apierror.InternalServerError(handler.app.Logger, w, r, err)
		return
	}

	if err := httputil.JSONResponse(w, http.StatusAccepted, doc, "file uploaded successfully"); err != nil {
		apierror.InternalServerError(handler.app.Logger, w, r, err)
	}
}
