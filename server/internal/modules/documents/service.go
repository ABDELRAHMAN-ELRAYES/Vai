package documents

import (
	"context"

	sharedDocuments "github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/shared/modules/documents"
	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) CreateDocument(ctx context.Context, ownerID uuid.UUID, uploadedFile *sharedDocuments.UploadedFile) (*Document, error) {
	doc := &Document{
		OwnerID:      ownerID,
		Name:         uploadedFile.FileName,
		OriginalName: uploadedFile.OriginalName,
		Size:         uploadedFile.Size,
		MimeType:     uploadedFile.MimeType,
	}

	if err := s.repo.Create(ctx, doc); err != nil {
		return nil, err
	}

	return doc, nil
}
