package documents

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/config"
	sharedDocuments "github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/shared/modules/documents"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/rag-engine/ai"
	apierror "github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/errors"
	"github.com/qdrant/go-client/qdrant"
	"go.uber.org/zap"

	"github.com/ABDELRAHMAN-ELRAYES/go-chunker"
	"github.com/google/uuid"
)
type IService interface {
	GetDocument(ctx context.Context, id string) (*Document, error)
	Search(ctx context.Context, query string, documentIDs []string, topK uint64) ([]string, error)
	EmbedDocument(ctx context.Context, documentID string, chunksDir string) error
}
type Service struct {
	repo      IRepository
	aiService ai.IService
	logger    *zap.SugaredLogger
}

func NewService(repo IRepository, aiService ai.IService, logger *zap.SugaredLogger) *Service {
	return &Service{
		repo:      repo,
		aiService: aiService,
		logger:    logger,
	}
}

func (service *Service) GetDocument(ctx context.Context, id string) (*Document, error) {
	return service.repo.GetDocumentByID(ctx, id)
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

func (service *Service) DeleteDocument(ctx context.Context, documentID string, uploadDir string, chunksDir string) error {
	// Get the document
	doc, err := service.repo.GetDocumentByID(ctx, documentID)
	if err != nil {
		return err
	}

	// Delete from Postgres
	err = service.repo.DeleteDocument(ctx, documentID)

	// Delete from Qdrant
	_ = service.repo.DeletePointsByDocumentID(ctx, documentID)

	// Delete File Storage
	if doc.Name != "" {
		_ = os.Remove(filepath.Join(uploadDir, doc.Name))
		ext := filepath.Ext(doc.Name)
		base := strings.TrimSuffix(doc.Name, ext)
		_ = os.Remove(filepath.Join(chunksDir, base+"_chunks.json"))
	}

	return err
}

// document status : draft -> processing -> ready / failed
func (service *Service) EmbedDocument(ctx context.Context, documentID string, chunksDir string) error {
	doc, err := service.repo.GetDocumentByID(ctx, documentID)
	if err != nil {
		return err
	}

	// Process the document if the status is draft
	if doc.Status != "draft" {
		return nil
	}

	// Set status to processing
	_ = service.repo.UpdateStatus(ctx, documentID, "processing")

	// Determine file name of the chunk
	ext := filepath.Ext(doc.Name)
	base := strings.TrimSuffix(doc.Name, ext)
	chunksFilePath := filepath.Join(chunksDir, base+"_chunks.json")

	// Read chunk JSON
	data, err := os.ReadFile(chunksFilePath)
	if err != nil {
		_ = service.repo.UpdateStatus(ctx, documentID, "failed")
		return apierror.ErrReadChunksFailed
	}

	var chunksFile DocumentChunksContent
	if err := json.Unmarshal(data, &chunksFile); err != nil {
		_ = service.repo.UpdateStatus(ctx, documentID, "failed")
		return apierror.ErrUnmarshalChunksFailed
	}

	var points []*qdrant.PointStruct

	var chunksModelInput []string

	for _, chunk := range chunksFile.Chunks {
		// Render prompt for the embedding model
		promptData := &EmbedPromptData{
			Text: chunk.Text,
		}
		prompt, err := ai.RenderPrompt(ai.EmbedDocumentPrompt, promptData)
		if err != nil {
			_ = service.repo.UpdateStatus(ctx, documentID, "failed")
			return apierror.ErrEmbedChunksFailed
		}
		chunksModelInput = append(chunksModelInput, prompt)
	}

	if len(chunksModelInput) == 0 {
		return service.repo.UpdateStatus(ctx, documentID, "ready")
	}
	// Embed the file chunks
	embeddings, err := service.aiService.EmbedBatch(ctx, chunksModelInput)
	if err != nil {
		_ = service.repo.UpdateStatus(ctx, documentID, "failed")
		return apierror.ErrEmbedChunksFailed
	}

	// Form the Qdrant points
	for i, chunk := range chunksFile.Chunks {
		embedding := embeddings[i]

		pointID := uuid.New().String()
		payload := map[string]any{
			"document_id": documentID,
			"text":        chunk.Text,
			"index":       chunk.Index,
			"start_char":  chunk.StartChar,
			"end_char":    chunk.EndChar,
		}

		points = append(points, &qdrant.PointStruct{
			Id:      qdrant.NewIDUUID(pointID),
			Vectors: qdrant.NewVectors(embedding...),
			Payload: qdrant.NewValueMap(payload),
		})
	}

	// Upsert vectors
	if len(points) > 0 {
		if err := service.repo.UpsertPoints(ctx, points); err != nil {
			_ = service.repo.UpdateStatus(ctx, documentID, "failed")
			return apierror.ErrUpsertVectorsFailed
		}
	}

	// Update status
	return service.repo.UpdateStatus(ctx, documentID, "ready")
}

func (service *Service) Search(ctx context.Context, query string, documentIDs []string, topK uint64) ([]string, error) {
	// 1. Render the search query prompt
	promptData := &EmbedPromptData{
		Text: query,
	}
	prompt, err := ai.RenderPrompt(ai.EmbedQueryPrompt, promptData)
	if err != nil {
		return nil, err
	}

	// 2. Embed the query
	embeddings, err := service.aiService.EmbedBatch(ctx, []string{prompt})
	if err != nil {
		return nil, err
	}

	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	// 3. Search in Qdrant
	points, err := service.repo.SearchPoints(ctx, embeddings[0], documentIDs, topK)
	if err != nil {
		return nil, err
	}

	// 4. Extract text from payloads
	var results []string
	for _, point := range points {
		if text, ok := point.Payload["text"]; ok {
			content := text.GetStringValue()
			results = append(results, content)
		}
	}

	return results, nil
}

func (service *Service) CleanupOldDrafts(ctx context.Context, uploadDir string, chunksDir string) error {
	// Find drafts older than 24 hours
	drafts, err := service.repo.GetOldDrafts(ctx, 24*time.Hour)
	if err != nil {
		return err
	}

	if len(drafts) == 0 {
		return nil
	}

	for _, doc := range drafts {
		if err := service.DeleteDocument(ctx, doc.ID.String(), uploadDir, chunksDir); err != nil {
			service.logger.Infow("Failed to delete document", "error", err)

			continue
		}
	}

	return nil
}
