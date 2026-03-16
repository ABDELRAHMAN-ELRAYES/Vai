package users

import (
	"context"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, handler *Handler) {
	fetchUser := func(ctx context.Context, id string) (any, error) {
		return handler.service.GetUser(ctx, id)
	}
	r.Route("/users", func(r chi.Router) {

		r.Use(middleware.Protect(handler.app, fetchUser))
		r.Post("/", handler.CreateUser)

		r.Get("/{userID}", handler.GetUser)
	})
}
