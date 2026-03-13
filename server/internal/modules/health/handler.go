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
		app: app,
	}
}

// healthCheckHandler godoc
//
//	@Summary		Health check
//	@Description	Returns service health and version information
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Failure		500	{object}	error
//	@Router			/health [get]
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":  "ok",
		"env":     h.app.Config.Env,
		"version": "0.0.1",
	}

	if err := httputil.JSONResponse(w, http.StatusOK, data, "This is the server Details"); err != nil {
		apierror.InternalServerError(h.app.Logger, w, r, err)
		return
	}

}
