package documents

import (
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, handler *Handler, getUser middleware.GetUser) {
	r.Route("/documents", func(r chi.Router) {
		r.Use(middleware.Protect(handler.app, getUser))

		r.With(middleware.FileUploadMiddleware(handler.app)).Post("/upload", handler.Upload)
	})
}
