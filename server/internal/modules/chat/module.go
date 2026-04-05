package chat

import (
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/middleware"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/users"
	"github.com/go-chi/chi/v5"
)

type Module struct {
	handler *Handler
	getUser middleware.GetUser
}

func New(
	app *app.Application,
	userService *users.Service,
	getUser middleware.GetUser) *Module {

	repo := NewRepository(app.DB)
	service := NewService(app.DB, repo, app.RAG.AI.Service, app.Logger, userService)
	handler := NewHandler(app, service)

	return &Module{
		handler: handler,
		getUser: getUser,
	}
}

func (m *Module) Name() string {
	return "conversations"
}
func (m *Module) RegisterRoutes(r chi.Router) {
	RegisterRoutes(r, m.handler, m.getUser)
}
