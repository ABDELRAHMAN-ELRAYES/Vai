package auth

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router, handler *Handler) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", handler.RegisterUser)
		r.Post("/activate/{token}", handler.ActivateUser)
		r.Post("/login", handler.AuthenticateUser)
	})
}
