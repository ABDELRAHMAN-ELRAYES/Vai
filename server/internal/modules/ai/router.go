package ai

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router, handler *Handler) {

	r.Route("/ai", func(r chi.Router) {
		r.HandleFunc("/generate", handler.Generate)
	})
}
