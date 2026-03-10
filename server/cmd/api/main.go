package main

import (
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/config"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/server"
	"go.uber.org/zap"
)

func main() {
	// Config Variables
	cfg := config.Load()

	// Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer func() {
		_ = logger.Sync()
	}()

	app := app.New(
		cfg,
		logger,
	)
	mux := server.NewRouter(app)
	logger.Fatal(server.Run(app, mux))

}
