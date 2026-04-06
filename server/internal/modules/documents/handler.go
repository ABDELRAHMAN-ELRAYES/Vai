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
	ctx := r.Context()
	ownerID, err := uuid.Parse(user.ID)
	if err != nil {
		apierror.InternalServerError(handler.app.Logger, w, r, err)
		return
	}

	// Save document metadata (status: draft)
	doc, err := handler.service.CreateDocument(ctx, ownerID, result)
	if err != nil {
		apierror.InternalServerError(handler.app.Logger, w, r, err)
		return
	}

	// Chunk document
	err = handler.service.GenerateChunks(&handler.app.Config.RAG.Chunker, handler.app.Config.Upload.Dir, doc)
	if err != nil {
		handler.app.Logger.Error("failed to generate chunks", "error", err)
		// delete the document if chunking failed
		err = handler.service.DeleteDocument(ctx, doc.ID.String(), handler.app.Config.Upload.Dir, handler.app.Config.RAG.Chunker.ChunksDir)
		if err != nil {
			handler.app.Logger.Error("failed to delete document", "error", err)
		}
		apierror.InternalServerError(handler.app.Logger, w, r, err)
		return
	}

	if err := httputil.JSONResponse(w, http.StatusAccepted, doc, "file uploaded successfully"); err != nil {
		apierror.InternalServerError(handler.app.Logger, w, r, err)
	}
}
