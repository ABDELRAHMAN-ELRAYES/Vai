package chat

import (
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/ai"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/users"
	"github.com/go-chi/chi/v5"
)

type Module struct {
	handler *Handler
}

func New(app *app.Application, aiService *ai.Service, userService *users.Service) *Module {
	repo := NewRepository(app.DB)
	service := NewService(app.DB, repo, aiService, app.Logger, userService)
	handler := NewHandler(app, service)

	return &Module{
		handler: handler,
	}
}

func (m *Module) Name() string {
	return "conversations"
}
func (m *Module) RegisterRoutes(r chi.Router) {
	RegisterRoutes(r, m.handler)
}
