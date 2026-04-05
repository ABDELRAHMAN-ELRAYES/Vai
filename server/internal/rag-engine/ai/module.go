package ai

import (
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/config"
	"go.uber.org/zap"
)

var (
	UserRole = "user"
	AIRole   = "ai"
)

type AIModule struct {
	Client  *Client
	Service *Service
}

func New(logger *zap.SugaredLogger, cfg *config.AI) *AIModule {

	client := NewClient(cfg, cfg.BaseURL)
	service := NewService(client)

	err := LoadPrompts()
	if err != nil {
		logger.Info("Prompts : ", err)
	}

	return &AIModule{
		Client:  client,
		Service: service,
	}
}
