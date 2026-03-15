package users

import (
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, handler *Handler) {

	r.Route("/users", func(r chi.Router) {
		r.Post("/", handler.CreateUser)
		r.Get("/{userID}", handler.GetUser)

	})
}
