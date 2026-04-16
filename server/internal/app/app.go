package app

import (
	"database/sql"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/auth"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/config"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/jobs"
	ratelimiter "github.com/ABDELRAHMAN-ELRAYES/Vai/internal/limiter"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/mailer"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/rag-engine"
	"github.com/qdrant/go-client/qdrant"
	"go.uber.org/zap"
)

type Application struct {
	Config        config.Config
	Logger        *zap.SugaredLogger
	DB            *sql.DB
	QdrantDB      *qdrant.Client
	Authenticator *auth.JWTAuthenticator
	Mailer        mailer.Client
	RAG           *rag.RAGEngine
	Scheduler     *jobs.Scheduler
	RateLimiter   ratelimiter.Limiter
}

func New(
	cfg config.Config,
	logger *zap.SugaredLogger,
	database *sql.DB,
	qdrantClient *qdrant.Client,
	authenticator *auth.JWTAuthenticator,
	mailer mailer.Client,
	rag *rag.RAGEngine,
	scheduler *jobs.Scheduler,
	rateLimiter ratelimiter.Limiter,
) *Application {
	return &Application{
		Config:        cfg,
		Logger:        logger,
		DB:            database,
		QdrantDB:      qdrantClient,
		Authenticator: authenticator,
		Mailer:        mailer,
		RAG:           rag,
		Scheduler:     scheduler,
		RateLimiter:   rateLimiter,
	}
}
