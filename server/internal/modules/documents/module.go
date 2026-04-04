package documents

import (
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/middleware"
	"github.com/go-chi/chi/v5"
)

type Module struct {
	handler     *Handler
	getUser middleware.GetUser
}

func New(app *app.Application, getUser middleware.GetUser) *Module {

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
		handler:     handler,
		getUser: getUser,
	}

}

func (m *Module) Name() string {
	return "documents"
}
func (m *Module) RegisterRoutes(r chi.Router) {
	RegisterRoutes(r, m.handler, m.getUser)

}
