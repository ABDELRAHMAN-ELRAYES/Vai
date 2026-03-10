package modules

import (
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/health"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/users"
	"github.com/go-chi/chi/v5"
)

func Register(r chi.Router, app *app.Application) {

	modules := []Module{
		users.New(app),
		health.New(app),
	}

	for _, m := range modules {
		m.RegisterRoutes(r)
	}
}
