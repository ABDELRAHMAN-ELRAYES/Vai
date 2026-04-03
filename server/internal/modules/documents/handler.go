package documents

import (
	"net/http"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"
	sharedDocs "github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/shared/modules/documents"
	apierror "github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/errors"
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

func (handler *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	result, ok := sharedDocs.GetUploadedFile(r)
	if !ok {
		apierror.InternalServerError(handler.app.Logger, w, r, apierror.ErrNoFileProvided)
		return
	}

	response := map[string]interface{}{
		"message":  "file uploaded successfully",
		"filename": result.FileName,
		"size":     result.Size,
	}

	if err := httputil.JSONResponse(w, http.StatusCreated, response, "file uploaded successfully"); err != nil {
		apierror.InternalServerError(handler.app.Logger, w, r, err)
	}
}
