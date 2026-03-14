package users

import (
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"
	"github.com/go-chi/chi/v5"
)

type Module struct {
	Handler *Handler
	Service *Service
}

func New(app *app.Application) *Module {

	repo := NewRepository(app.DB)
	service := NewService(repo)
	handler := NewHandler(app, service)

	return &Module{
		Handler: handler,
		Service: service,
	}
}

func (m *Module) Name() string {
	return "users"
}

func (m *Module) RegisterRoutes(r chi.Router) {
	RegisterRoutes(r, m.Handler)
}
