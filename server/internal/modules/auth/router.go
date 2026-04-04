package auth

import (
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, handler *Handler, getUser middleware.GetUser) {

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", handler.RegisterUser)
		r.Post("/activate/{token}", handler.ActivateUser)
		r.Post("/login", handler.AuthenticateUser)
		r.Post("/logout", handler.Logout)
		r.With(middleware.Protect(handler.app, getUser)).Get("/me", handler.GetMe)
	})
}
