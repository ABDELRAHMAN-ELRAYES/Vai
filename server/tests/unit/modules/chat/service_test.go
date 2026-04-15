package chat_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/chat"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/documents"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/rag-engine/ai"
	apierror "github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// Mocks
type mockChatRepo struct {
	mock.Mock
}

func (m *mockChatRepo) CreateConversation(ctx context.Context, conv *chat.Conversation) error {
	args := m.Called(ctx, conv)
	return args.Error(0)
}

func (m *mockChatRepo) CreateMessage(ctx context.Context, msg *chat.Message) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *mockChatRepo) UpdateConversation(ctx context.Context, conv *chat.Conversation) error {
	args := m.Called(ctx, conv)
	return args.Error(0)
}

func (m *mockChatRepo) GetConversationsByUserID(ctx context.Context, userID string) ([]*chat.Conversation, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*chat.Conversation), args.Error(1)
}

func (m *mockChatRepo) DeleteConversation(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockChatRepo) GetConversationByID(ctx context.Context, id string) (*chat.Conversation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*chat.Conversation), args.Error(1)
}

func (m *mockChatRepo) GetMessagesByConversationID(ctx context.Context, id string) ([]chat.Message, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]chat.Message), args.Error(1)
}

func (m *mockChatRepo) AddMessageDocuments(ctx context.Context, msgID string, docIDs []string) error {
	args := m.Called(ctx, msgID, docIDs)
	return args.Error(0)
}

func (m *mockChatRepo) GetAssociatedDocumentIDs(ctx context.Context, convID string) ([]string, error) {
	args := m.Called(ctx, convID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockChatRepo) WithTx(tx *sql.Tx) chat.IRepository {
	return m
}

type mockDocsService struct {
	mock.Mock
}

func (m *mockDocsService) GetDocument(ctx context.Context, id string) (*documents.Document, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*documents.Document), args.Error(1)
}

func (m *mockDocsService) Search(ctx context.Context, query string, docIDs []string, topK uint64) ([]string, error) {
	args := m.Called(ctx, query, docIDs, topK)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockDocsService) EmbedDocument(ctx context.Context, id string, dir string) error {
	args := m.Called(ctx, id, dir)
	return args.Error(0)
}

type mockAIService struct {
	mock.Mock
}

func (m *mockAIService) Generate(ctx context.Context, prompt string) (<-chan string, <-chan error) {
	args := m.Called(ctx, prompt)
	return args.Get(0).(<-chan string), args.Get(1).(<-chan error)
}

func (m *mockAIService) CollectTokens(tChan <-chan string, eChan <-chan error) (<-chan string, <-chan string, <-chan error) {
	args := m.Called(tChan, eChan)
	return args.Get(0).(<-chan string), args.Get(1).(<-chan string), args.Get(2).(<-chan error)
}

func (m *mockAIService) EmbedBatch(ctx context.Context, input []string) ([][]float32, error) {
	args := m.Called(ctx, input)
	return args.Get(0).([][]float32), args.Error(1)
}

func TestChatService_GetConversations(t *testing.T) {
	db, _, _ := sqlmock.New()
	mockRepo := new(mockChatRepo)
	logger := zap.NewNop().Sugar()
	service := chat.NewService(db, mockRepo, nil, logger, nil, nil)

	t.Run("successful list conversations", func(t *testing.T) {
		userID := "user-123"
		expected := []*chat.Conversation{{ID: "conv-1", Title: "Test"}}

		mockRepo.On("GetConversationsByUserID", mock.Anything, userID).Return(expected, nil)

		res, err := service.GetConversations(context.Background(), userID)

		assert.NoError(t, err)
		assert.Equal(t, expected, res)
		mockRepo.AssertExpectations(t)
	})
}

func TestChatService_StartConversation(t *testing.T) {
	db, mockDB, _ := sqlmock.New()
	mockRepo := new(mockChatRepo)
	mockDocs := new(mockDocsService)
	mockAI := new(mockAIService)
	logger := zap.NewNop().Sugar()

	// Ensure prompts are loaded for RenderPrompt
	_ = ai.LoadPrompts()

	service := chat.NewService(db, mockRepo, mockAI, logger, nil, mockDocs)

	ctx := context.Background()
	payload := chat.StartConversationPayload{
		UserID:      "user-1",
		Title:       "Test Conv",
		Message:     "Hello?",
		DocumentIDs: []string{"doc-1"},
	}

	t.Run("successful start conversation", func(t *testing.T) {
		mockDB.ExpectBegin()
		mockDB.ExpectCommit()

		// helper: embedDocuments
		mockDocs.On("EmbedDocument", mock.Anything, "doc-1", "dir").Return(nil)

		// repo calls
		mockRepo.On("CreateConversation", mock.Anything, mock.AnythingOfType("*chat.Conversation")).
			Run(func(args mock.Arguments) {
				conv := args.Get(1).(*chat.Conversation)
				conv.ID = "conv-1"
			}).
			Return(nil)

		mockRepo.On("CreateMessage", mock.Anything, mock.MatchedBy(func(msg *chat.Message) bool {
			return msg.Role == ai.UserRole && msg.Content == payload.Message
		})).Return(nil)

		mockRepo.On("AddMessageDocuments", mock.Anything, mock.Anything, payload.DocumentIDs).Return(nil)

		// helper: getSemanticContext
		mockDocs.On("GetDocument", mock.Anything, "doc-1").Return(&documents.Document{OriginalName: "test.pdf"}, nil)
		mockDocs.On("Search", mock.Anything, payload.Message, payload.DocumentIDs, uint64(5)).Return([]string{"chunk1"}, nil)

		// helper: streamAIResponse
		tokenChan := make(chan string, 1)
		errChan := make(chan error, 1)
		mockAI.On("Generate", mock.Anything, mock.AnythingOfType("string")).Return((<-chan string)(tokenChan), (<-chan error)(errChan))

		replyChan := make(chan string, 1)
		tokenStream := make(chan string, 1)
		errStream := make(chan error, 1)
		mockAI.On("CollectTokens", mock.Anything, mock.Anything).Return((<-chan string)(replyChan), (<-chan string)(tokenStream), (<-chan error)(errStream))

		conv, tStream, eStream, err := service.StartConversation(ctx, payload, "dir")

		assert.NoError(t, err)
		assert.NotNil(t, conv)
		assert.Equal(t, "conv-1", conv.ID)
		assert.NotNil(t, tStream)
		assert.NotNil(t, eStream)

		mockRepo.AssertExpectations(t)
		mockDocs.AssertExpectations(t)
		mockAI.AssertExpectations(t)
		assert.NoError(t, mockDB.ExpectationsWereMet())
	})

	t.Run("validation error", func(t *testing.T) {
		invalidPayload := chat.StartConversationPayload{} // Missing required fields
		_, _, _, err := service.StartConversation(ctx, invalidPayload, "dir")
		assert.ErrorIs(t, err, apierror.ErrBadRequest)
	})
}

func TestChatService_SendMessage(t *testing.T) {
	db, mockDB, _ := sqlmock.New()
	mockRepo := new(mockChatRepo)
	mockDocs := new(mockDocsService)
	mockAI := new(mockAIService)
	logger := zap.NewNop().Sugar()

	_ = ai.LoadPrompts()

	service := chat.NewService(db, mockRepo, mockAI, logger, nil, mockDocs)

	ctx := context.Background()
	payload := chat.SendMessagePayload{
		ConversationID: "conv-1",
		UserID:         "user-1",
		Message:        "Tell me about this document",
		DocumentIDs:    []string{"doc-2"},
	}

	conv := &chat.Conversation{ID: "conv-1", UserID: "user-1"}

	setupCommonMocks := func() {
		mockRepo.On("GetConversationByID", mock.Anything, "conv-1").Return(conv, nil)
		mockDocs.On("EmbedDocument", mock.Anything, "doc-2", "dir").Return(nil)

		mockDB.ExpectBegin()
		mockRepo.On("CreateMessage", mock.Anything, mock.AnythingOfType("*chat.Message")).Return(nil)
		mockRepo.On("AddMessageDocuments", mock.Anything, mock.Anything, payload.DocumentIDs).Return(nil)
		mockDB.ExpectCommit()

		mockRepo.On("GetMessagesByConversationID", mock.Anything, "conv-1").Return([]chat.Message{}, nil)

		// Mock AI response
		tokenChan := make(chan string, 1)
		errChan := make(chan error, 1)
		mockAI.On("Generate", mock.Anything, mock.Anything).Return((<-chan string)(tokenChan), (<-chan error)(errChan))

		replyChan := make(chan string, 1)
		tokenStream := make(chan string, 1)
		errStream := make(chan error, 1)
		mockAI.On("CollectTokens", mock.Anything, mock.Anything).Return((<-chan string)(replyChan), (<-chan string)(tokenStream), (<-chan error)(errStream))
	}

	t.Run("successful send message - message scoped", func(t *testing.T) {
		setupCommonMocks()

		// Message scoped search because of "this document" in message
		mockDocs.On("GetDocument", mock.Anything, "doc-2").Return(&documents.Document{OriginalName: "doc2.pdf"}, nil)
		mockDocs.On("Search", mock.Anything, payload.Message, []string{"doc-2"}, uint64(5)).Return([]string{"found context"}, nil)

		tStream, eStream, err := service.SendMessage(ctx, payload, "dir")

		assert.NoError(t, err)
		assert.NotNil(t, tStream)
		assert.NotNil(t, eStream)

		mockRepo.AssertExpectations(t)
		mockDocs.AssertExpectations(t)
		assert.NoError(t, mockDB.ExpectationsWereMet())
	})

	t.Run("successful send message - conversation scoped", func(t *testing.T) {
		// Reset mocks for this run
		mockRepo = new(mockChatRepo)
		mockDocs = new(mockDocsService)
		mockAI = new(mockAIService)
		service = chat.NewService(db, mockRepo, mockAI, logger, nil, mockDocs)

		payloadConvScoped := payload
		payloadConvScoped.Message = "Hello" // Not message scoped

		mockRepo.On("GetConversationByID", mock.Anything, "conv-1").Return(conv, nil)
		mockDocs.On("EmbedDocument", mock.Anything, "doc-2", "dir").Return(nil)

		mockDB.ExpectBegin()
		mockRepo.On("CreateMessage", mock.Anything, mock.AnythingOfType("*chat.Message")).Return(nil)
		mockRepo.On("AddMessageDocuments", mock.Anything, mock.Anything, payload.DocumentIDs).Return(nil)
		mockDB.ExpectCommit()

		mockRepo.On("GetMessagesByConversationID", mock.Anything, "conv-1").Return([]chat.Message{}, nil)

		// Conversation scoped IDs
		mockRepo.On("GetAssociatedDocumentIDs", mock.Anything, "conv-1").Return([]string{"doc-1", "doc-2"}, nil)
		mockDocs.On("GetDocument", mock.Anything, "doc-1").Return(&documents.Document{OriginalName: "doc1.pdf"}, nil)
		mockDocs.On("GetDocument", mock.Anything, "doc-2").Return(&documents.Document{OriginalName: "doc2.pdf"}, nil)
		mockDocs.On("Search", mock.Anything, "Hello", []string{"doc-1", "doc-2"}, uint64(5)).Return([]string{"conv context"}, nil)

		// Mock AI response
		mockAI.On("Generate", mock.Anything, mock.Anything).Return((<-chan string)(make(chan string)), (<-chan error)(make(chan error)))
		mockAI.On("CollectTokens", mock.Anything, mock.Anything).Return((<-chan string)(make(chan string)), (<-chan string)(make(chan string)), (<-chan error)(make(chan error)))

		_, _, err := service.SendMessage(ctx, payloadConvScoped, "dir")

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockDocs.AssertExpectations(t)
	})

	t.Run("unauthorized", func(t *testing.T) {
		mockRepo = new(mockChatRepo)
		service = chat.NewService(db, mockRepo, mockAI, logger, nil, mockDocs)

		otherUserConv := &chat.Conversation{ID: "conv-1", UserID: "other-user"}
		mockRepo.On("GetConversationByID", mock.Anything, "conv-1").Return(otherUserConv, nil)

		_, _, err := service.SendMessage(ctx, payload, "dir")

		assert.ErrorIs(t, err, apierror.ErrUnauthorized)
	})

	t.Run("conversation not found", func(t *testing.T) {
		mockRepo = new(mockChatRepo)
		service = chat.NewService(db, mockRepo, mockAI, logger, nil, mockDocs)

		mockRepo.On("GetConversationByID", mock.Anything, "conv-1").Return(nil, sql.ErrNoRows)

		_, _, err := service.SendMessage(ctx, payload, "dir")

		assert.ErrorIs(t, err, sql.ErrNoRows)
	})
}
