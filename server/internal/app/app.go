package app

import (
	"database/sql"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/auth"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/config"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/mailer"
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
}

func New(cfg config.Config,
	logger *zap.SugaredLogger,
	database *sql.DB,
	qdrantClient *qdrant.Client,
	authenticator *auth.JWTAuthenticator,
	mailer mailer.Client) *Application {
	return &Application{
		Config:        config.Load(),
		Logger:        logger,
		DB:            database,
		QdrantDB:      qdrantClient,
		Authenticator: authenticator,
		Mailer:        mailer,
	}
}
