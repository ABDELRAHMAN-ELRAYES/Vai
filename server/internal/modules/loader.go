package modules

import (
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/ai"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/auth"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/documents"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/health"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/users"
	"github.com/go-chi/chi/v5"
)

func Register(r chi.Router, app *app.Application) {

	userModule := users.New(app)

	modules := []Module{
		health.New(app),
		userModule,
		ai.New(app),
		documents.New(app),
		auth.New(app, userModule.Service),
	}

	for _, m := range modules {
		m.RegisterRoutes(r)
	}
}
