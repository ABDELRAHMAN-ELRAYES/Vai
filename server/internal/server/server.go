package server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	docs "github.com/ABDELRAHMAN-ELRAYES/Vai/docs/api"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"
	"github.com/go-chi/chi/v5"
)

const version = "0.0.1"

// Run starts the HTTP server with graceful shutdown.
func Run(app *app.Application, mux *chi.Mux) error {

	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = app.Config.APIURL
	docs.SwaggerInfo.BasePath = "/api/v1"

	server := &http.Server{
		Addr:         app.Config.Addr,
		Handler:      mux,
		WriteTimeout: 15 * time.Minute,
		ReadTimeout:  15 * time.Minute,
		IdleTimeout:  time.Minute,
	}

	// graceful shutdown
	shutdown := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		sig := <-quit

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		app.Logger.Infow("signal caught", "signal", sig.String())
		shutdown <- server.Shutdown(ctx)
	}()

	app.Logger.Infow("server has started", "addr", app.Config.Addr, "env", app.Config.Env)

	// Start Background Job Scheduler
	app.Scheduler.Start()
	defer app.Scheduler.Stop()

	err := server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdown
	if err != nil {
		return err
	}

	app.Logger.Infow("server has stopped", "addr", app.Config.Addr, "env", app.Config.Env)

	return nil
}
