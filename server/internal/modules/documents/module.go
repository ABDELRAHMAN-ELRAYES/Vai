package documents

import (
	"time"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/config"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/jobs"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/middleware"
	"github.com/go-chi/chi/v5"
)

type Module struct {
	handler *Handler
	Service *Service
	getUser middleware.GetUser
	cfg     *config.Config
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

	service := NewService(repo, app.RAG.AI.Service, app.Logger)
	handler := NewHandler(app, service)

	return &Module{
		handler: handler,
		Service: service,
		getUser: getUser,
		cfg:     &app.Config,
	}

}

func (m *Module) Name() string {
	return "documents"
}
func (m *Module) RegisterRoutes(r chi.Router) {
	RegisterRoutes(r, m.handler, m.getUser)

}

func (m *Module) RegisterJobs(scheduler *jobs.Scheduler) {
	cleanupJob := NewCleanupDraftsJob(m.Service, m.cfg)
	scheduler.Register(cleanupJob, 12*time.Hour)
}
