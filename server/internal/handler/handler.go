package handler

import "github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"

type Handler struct {
	App *app.Application
}

func New(app *app.Application) *Handler {
	return &Handler{App: app}
}
