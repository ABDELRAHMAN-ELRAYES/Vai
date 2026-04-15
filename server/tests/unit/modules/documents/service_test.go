package documents_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/documents"
	sharedDocuments "github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/shared/modules/documents"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/rag-engine/ai"
	"github.com/google/uuid"
	"github.com/qdrant/go-client/qdrant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	if err := ai.LoadPrompts(); err != nil {
		fmt.Printf("failed to load prompts: %v\n", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

// Mocks
type mockDocsRepo struct {
	mock.Mock
}

func (m *mockDocsRepo) Create(ctx context.Context, doc *documents.Document) error {
	args := m.Called(ctx, doc)
	return args.Error(0)
}

func (m *mockDocsRepo) GetDocumentByID(ctx context.Context, id string) (*documents.Document, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*documents.Document), args.Error(1)
}

func (m *mockDocsRepo) UpdateStatus(ctx context.Context, id string, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *mockDocsRepo) DeleteDocument(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockDocsRepo) UpsertPoints(ctx context.Context, points []*qdrant.PointStruct) error {
	args := m.Called(ctx, points)
	return args.Error(0)
}

func (m *mockDocsRepo) DeletePointsByDocumentID(ctx context.Context, documentID string) error {
	args := m.Called(ctx, documentID)
	return args.Error(0)
}

func (m *mockDocsRepo) SearchPoints(ctx context.Context, vector []float32, documentIDs []string, topK uint64) ([]*qdrant.ScoredPoint, error) {
	args := m.Called(ctx, vector, documentIDs, topK)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*qdrant.ScoredPoint), args.Error(1)
}

func (m *mockDocsRepo) GetOldDrafts(ctx context.Context, olderThan time.Duration) ([]*documents.Document, error) {
	args := m.Called(ctx, olderThan)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*documents.Document), args.Error(1)
}

type mockAIService struct {
	mock.Mock
}

func (m *mockAIService) Generate(ctx context.Context, prompt string) (<-chan string, <-chan error) {
	args := m.Called(ctx, prompt)
	var r0 <-chan string
	if args.Get(0) != nil {
		r0 = args.Get(0).(<-chan string)
	}
	var r1 <-chan error
	if args.Get(1) != nil {
		r1 = args.Get(1).(<-chan error)
	}
	return r0, r1
}

func (m *mockAIService) CollectTokens(tokenChan <-chan string, errChan <-chan error) (<-chan string, <-chan string, <-chan error) {
	args := m.Called(tokenChan, errChan)
	var r0 <-chan string
	if args.Get(0) != nil {
		r0 = args.Get(0).(<-chan string)
	}
	var r1 <-chan string
	if args.Get(1) != nil {
		r1 = args.Get(1).(<-chan string)
	}
	var r2 <-chan error
	if args.Get(2) != nil {
		r2 = args.Get(2).(<-chan error)
	}
	return r0, r1, r2
}

func (m *mockAIService) EmbedBatch(ctx context.Context, input []string) ([][]float32, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([][]float32), args.Error(1)
}

func TestDocumentsService_CreateDocument(t *testing.T) {
	mockRepo := &mockDocsRepo{}
	logger := zap.NewNop().Sugar()
	service := documents.NewService(mockRepo, nil, logger)

	t.Run("successful document creation", func(t *testing.T) {
		ownerID := uuid.New()
		uploadedFile := &sharedDocuments.UploadedFile{
			FileName:     "test.txt",
			OriginalName: "original.txt",
			Size:         100,
			MimeType:     "text/plain",
		}

		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*documents.Document")).Return(nil)

		doc, err := service.CreateDocument(context.Background(), ownerID, uploadedFile)

		assert.NoError(t, err)
		assert.NotNil(t, doc)
		assert.Equal(t, uploadedFile.FileName, doc.Name)
		assert.Equal(t, ownerID, doc.OwnerID)
		mockRepo.AssertExpectations(t)
	})
}

func TestDocumentsService_Search(t *testing.T) {
	mockRepo := &mockDocsRepo{}
	mockAI := &mockAIService{}
	logger := zap.NewNop().Sugar()
	service := documents.NewService(mockRepo, mockAI, logger)

	t.Run("successful search", func(t *testing.T) {
		query := "test query"
		docIDs := []string{uuid.New().String()}
		topK := uint64(5)

		mockAI.On("EmbedBatch", mock.Anything, mock.Anything).Return([][]float32{{0.1, 0.2}}, nil)
		
		scoredPoints := []*qdrant.ScoredPoint{
			{
				Payload: map[string]*qdrant.Value{
					"text": qdrant.NewValueString("result text"),
				},
			},
		}
		mockRepo.On("SearchPoints", mock.Anything, mock.Anything, docIDs, topK).Return(scoredPoints, nil)

		results, err := service.Search(context.Background(), query, docIDs, topK)

		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "result text", results[0])
		mockAI.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}
