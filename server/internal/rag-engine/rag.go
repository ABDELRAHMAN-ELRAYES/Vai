package rag

import (
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/config"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/rag-engine/ai"
	"go.uber.org/zap"
)

type RAGEngine struct {
	AI      *ai.AIModule
}

func New(logger *zap.SugaredLogger, cfg *config.RAGConfig) *RAGEngine {
	ai := ai.New(logger, &cfg.AI)

	return &RAGEngine{
		AI:      ai,
	}
}
