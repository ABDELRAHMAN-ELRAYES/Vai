package health

import (
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/jobs"
	"github.com/go-chi/chi/v5"
)

type Module struct {
	handler *Handler
}

func New(app *app.Application) *Module {

	handler := NewHandler(app)

	return &Module{
		handler: handler,
	}
}

func (m *Module) Name() string {
	return "health"
}

func (m *Module) RegisterRoutes(r chi.Router) {
	RegisterRoutes(r, m.handler)
}

func (m *Module) RegisterJobs(scheduler *jobs.Scheduler) {}
