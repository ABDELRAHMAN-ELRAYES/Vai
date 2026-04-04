package ai

import (
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"
)

var (
	UserRole = "user"
	AIRole   = "ai"
)

type Module struct {
	Client  *Client
	Service *Service
}

func New(app *app.Application) *Module {

	client := NewClient(app, app.Config.AI.BaseURL)
	service := NewService(client)

	err := LoadPrompts()
	if err != nil {
		app.Logger.Info("Prompts : ", err)
	}

	return &Module{
		Client:  client,
		Service: service,
	}
}
