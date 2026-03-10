package main

import (
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/config"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/db"
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

	database, err := db.New(
		cfg.DB.Addr,
		cfg.DB.MaxOpenConns,
		cfg.DB.MaxIdleConns,
		cfg.DB.MaxIdleTime,
	)

	if err != nil {
		logger.Fatal(err)
	}

	defer database.Close()
	logger.Info("Database connection pool established")

	app := app.New(
		cfg,
		logger,
		database,
	)
	// Create Router
	mux := server.NewRouter(app)

	// Run the server
	logger.Fatal(server.Run(app, mux))

}
