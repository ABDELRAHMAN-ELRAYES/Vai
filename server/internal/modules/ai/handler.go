package ai

import (
	"net/http"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"
)

type Handler struct {
	app     *app.Application
	service *Service
}

func NewHandler(app *app.Application, service *Service) *Handler {
	return &Handler{
		app:     app,
		service: service,
	}
}

func (h *Handler) Generate(w http.ResponseWriter, r *http.Request) {

	prompt := r.URL.Query().Get("prompt")

	ctx := r.Context()

	stream, errs := h.service.Generate(ctx, prompt)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Ensure to send each token immediately once it was recieved
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	for {
		select {

		case token, ok := <-stream:

			if !ok {
				return
			}

			_, _ = w.Write([]byte("data: " + token + "\n\n"))
			flusher.Flush()

		case err := <-errs:

			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}

		case <-ctx.Done():
			return
		}
	}
}
