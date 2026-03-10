package users

import (
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"
	"github.com/go-chi/chi/v5"
)

type Module struct {
	handler *Handler
}

func New(app *app.Application) *Module {

	repo := NewRepository(app.DB)
	service := NewService(repo)
	handler := NewHandler(app, service)

	return &Module{
		handler: handler,
	}
}

func (m *Module) Name() string {
	return "users"
}

func (m *Module) RegisterRoutes(r chi.Router) {
	RegisterRoutes(r, m.handler)
}
