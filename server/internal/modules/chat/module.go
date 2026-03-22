package chat

import "github.com/go-chi/chi/v5"

type Module struct {
}

func New() *Module {

	return &Module{}
}

func (m *Module) Name() string {
	return "converstions"
}
func (m *Module) RegisterRoutes(r chi.Router) {
	RegisterRoutes(r)
}
