package documents

import (
	"context"
	"fmt"
	"os"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/config"
	sharedDocuments "github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/shared/modules/documents"

	"github.com/ABDELRAHMAN-ELRAYES/go-chunker"
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

func (service *Service) CreateDocument(ctx context.Context, ownerID uuid.UUID, uploadedFile *sharedDocuments.UploadedFile) (*Document, error) {
	doc := &Document{
		OwnerID:      ownerID,
		Name:         uploadedFile.FileName,
		OriginalName: uploadedFile.OriginalName,
		Size:         uploadedFile.Size,
		MimeType:     uploadedFile.MimeType,
	}

	if err := service.repo.Create(ctx, doc); err != nil {
		return nil, err
	}

	return doc, nil
}

func (service *Service) GenerateChunks(cfg *config.ChunkerConfig, uploadDir string, doc *Document) error {

	source := fmt.Sprintf("%s/%s", uploadDir, doc.Name)
	// 1. Read the raw file
	content, err := os.ReadFile(source)
	if err != nil {
		return err
	}

	// 2. Initialize a strategy
	splitter := chunker.NewSplitter(
		chunker.WithSize(cfg.ChunkSize),
		chunker.WithOverlap(cfg.Overlap),
	)

	// 3. Split the text
	meta := chunker.Meta{
		DocumentID: doc.ID.String(),
		Source:     source,
	}
	chunks, err := splitter.Split(context.Background(), string(content), meta)
	if err != nil {
		return err
	}

	// 4. Write to JSON
	_, err = chunker.WriteJSON(chunks, source, meta.DocumentID, cfg.ChunksDir, splitter.Config())
	if err != nil {
		return err
	}

	return nil
}
