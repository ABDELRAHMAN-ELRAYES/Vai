package app

import (
	"database/sql"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/config"
	"github.com/qdrant/go-client/qdrant"
	"go.uber.org/zap"
)

type Application struct {
	Config   config.Config
	Logger   *zap.SugaredLogger
	DB       *sql.DB
	QdrantDB *qdrant.Client
}

func New(cfg config.Config, logger *zap.SugaredLogger, database *sql.DB, qdrantClient *qdrant.Client) *Application {
	return &Application{
		Config:   config.Load(),
		Logger:   logger,
		DB:       database,
		QdrantDB: qdrantClient,
	}
}
