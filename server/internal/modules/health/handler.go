package health

import (
	"net/http"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/apierror"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/httputil"
)

type Handler struct {
	app *app.Application
}

func NewHandler(app *app.Application) *Handler {
	return &Handler{
		app:     app,
	}
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":  "ok",
		"env":     h.app.Config.Env,
		"version": "0.0.1",
	}

	if err := httputil.JSONResponse(w, http.StatusOK, data); err != nil {
		apierror.InternalServerError(h.app.Logger, w, r, err)
		return
	}

}
