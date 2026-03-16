package auth

import (
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/users"
	"github.com/go-chi/chi/v5"
)

type Module struct {
	Handler *Handler
	Service *Service
}

func New(app *app.Application, userService *users.Service) *Module {
	repo := NewRepository(app.DB)
	service := NewService(app.DB, repo, userService, app.Authenticator, &app.Config, app.Mailer)
	handler := NewHandler(app, service)
	return &Module{
		Handler: handler,
		Service: service,
	}
}

func (module *Module) Name() string {
	return "Authentication"
}
func (module *Module) RegisterRoutes(r chi.Router) {
	RegisterRoutes(r, module.Handler)
}
