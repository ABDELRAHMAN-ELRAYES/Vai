package chat

import (
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, handler *Handler, getUser middleware.GetUser) {
	r.Route("/conversations", func(r chi.Router) {
		r.Use(middleware.Protect(handler.app, getUser))
		r.Post("/", handler.StartConversation)
		r.Get("/", handler.GetConversations)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", handler.GetChat)
			r.Post("/", handler.SendMessage)
			r.Patch("/", handler.UpdateConversation)
			r.Delete("/", handler.DeleteConversation)
		})
	})
}
