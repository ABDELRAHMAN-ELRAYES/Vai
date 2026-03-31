package chat

import (
	"context"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, handler *Handler) {
	fetchUser := func(ctx context.Context, id string) (any, error) {
		return handler.service.userService.GetUser(ctx, id)
	}

	r.Route("/conversations", func(r chi.Router) {
		r.Use(middleware.Protect(handler.app, fetchUser))
		r.Post("/", handler.StartConversation)
		r.Get("/", handler.GetConversations)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", handler.GetChat)
			r.Patch("/", handler.UpdateConversation)
			r.Delete("/", handler.DeleteConversation)
		})
	})
}
