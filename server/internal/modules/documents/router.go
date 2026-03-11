package documents

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router, handler *Handler) {
	r.Route("/documents", func(r chi.Router) {
		// TODO: Register all documents endpoints
	})
}
