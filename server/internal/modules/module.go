package modules

import (
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/jobs"
	"github.com/go-chi/chi/v5"
)

type Module interface {
	Name() string
	RegisterRoutes(r chi.Router)
	RegisterJobs(scheduler *jobs.Scheduler)
}
