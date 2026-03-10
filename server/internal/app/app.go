package app

import (
	"database/sql"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/config"
	"go.uber.org/zap"
)

type Application struct {
	Config config.Config
	Logger *zap.SugaredLogger
	DB     *sql.DB
}

func New(cfg config.Config, logger *zap.SugaredLogger, database *sql.DB) *Application {
	return &Application{
		Config: config.Load(),
		Logger: logger,
		DB:     database,
	}
}
