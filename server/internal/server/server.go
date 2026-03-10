package server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"
	"github.com/go-chi/chi/v5"
)

func Run(app *app.Application, mux *chi.Mux) error {
	server := &http.Server{
		Addr:         app.Config.Addr,
		Handler:      mux,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  10 * time.Second,
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
