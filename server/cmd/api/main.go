package main

import (
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/config"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/db"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/server"
	"go.uber.org/zap"
)

//	@title			Vai API
//	@version		1.0
//	@description	This is Vai Server
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		vai.swagger.io
//	@BasePath	/api/v1

// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description
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

	qdrantClient, err := db.NewQdrantClient(cfg.QdrantDB.Host, cfg.QdrantDB.Port)
	if err != nil {
		logger.Fatal(err)
	}

	defer qdrantClient.Close()
	logger.Info("Qdrant Database connection pool established")

	app := app.New(
		cfg,
		logger,
		database,
		qdrantClient,
	)
	// Create Router
	mux := server.NewRouter(app)

	// Run the server
	logger.Fatal(server.Run(app, mux))

}
