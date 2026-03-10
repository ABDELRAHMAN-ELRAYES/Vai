package app

import (
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/config"
	"go.uber.org/zap"
)

type Application struct {
	Config config.Config
	Logger *zap.SugaredLogger
}

func New(cfg config.Config, logger *zap.SugaredLogger) *Application {
	return &Application{
		Config: config.Load(),
		Logger: logger,
	}
}
