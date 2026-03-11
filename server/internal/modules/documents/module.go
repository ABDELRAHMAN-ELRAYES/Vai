package documents

import (
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"
	"github.com/go-chi/chi/v5"
)

type Module struct {
	handler *Handler
}

func New(app *app.Application) *Module {

	qdrantClient := &QdrantClient{
		client:         app.QdrantDB,
		collectionName: "documents",
	}

	repo := NewRepository(app.DB, qdrantClient)
	// Initialize the document collection
	if err := repo.InitCollection(); err != nil {
		app.Logger.Error(err)
		return nil
	}

	service := NewService(repo)
	handler := NewHandler(app, service)

	return &Module{
		handler: handler,
	}

}

func (m *Module) Name() string {
	return "documents"
}
func (m *Module) RegisterRoutes(r chi.Router) {
	RegisterRoutes(r, m.handler)

}
