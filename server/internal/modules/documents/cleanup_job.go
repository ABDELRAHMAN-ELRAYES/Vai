package documents

import (
	"context"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/config"
)


func NewCleanupDraftsJob(service *Service, cfg *config.Config) *CleanupDraftsJob {
	return &CleanupDraftsJob{
		service:   service,
		uploadDir: cfg.Upload.Dir,
		chunksDir: cfg.RAG.Chunker.ChunksDir,
	}
}

func (job *CleanupDraftsJob) Name() string {
	return "document-cleanup-drafts"
}

func (job *CleanupDraftsJob) Run(ctx context.Context) error {
	return job.service.CleanupOldDrafts(ctx, job.uploadDir, job.chunksDir)
}
