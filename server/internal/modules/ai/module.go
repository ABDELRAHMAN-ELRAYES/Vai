package ai

import (
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"
	"github.com/go-chi/chi/v5"
)

var (
	UserRole = "user"
	AIRole   = "ai"
)

type Module struct {
	Service *Service
	handler *Handler
}

func New(app *app.Application) *Module {

	client := NewClient(app, app.Config.AI.BaseURL)

	service := NewService(client)
	handler := NewHandler(app, service)

	err := LoadPrompts()
	if err != nil {
		app.Logger.Info("Prompts : ", err)
	}

	return &Module{
		Service: service,
		handler: handler,
	}
}

func (m *Module) Name() string {
	return "ai"
}

func (m *Module) RegisterRoutes(r chi.Router) {
	RegisterRoutes(r, m.handler)
}
