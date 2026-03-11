package documents

import "github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"

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
