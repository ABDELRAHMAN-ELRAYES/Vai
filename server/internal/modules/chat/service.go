package chat

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/db"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/documents"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/users"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/rag-engine/ai"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/validator"
	apierror "github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/errors"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/utils"
	"go.uber.org/zap"
)

type Service struct {
	db          *sql.DB
	repo        *Repository
	aiService   *ai.Service
	logger      *zap.SugaredLogger
	userService *users.Service
	docService  *documents.Service
}

func NewService(db *sql.DB, repo *Repository, aiService *ai.Service, logger *zap.SugaredLogger, userService *users.Service, docService *documents.Service) *Service {
	return &Service{
		db:          db,
		repo:        repo,
		aiService:   aiService,
		logger:      logger,
		userService: userService,
		docService:  docService,
	}
}
func (service *Service) StartConversation(ctx context.Context, payload StartConversationPayload, chunksDir string) (*Conversation, <-chan string, <-chan error, error) {
	// Validate request body
	if err := validator.Validate.Struct(payload); err != nil {
		return nil, nil, nil, apierror.ErrBadRequest
	}

	conv := &Conversation{}

	// Embed all documents if provided with the message
	for _, docID := range payload.DocumentIDs {
		err := service.docService.EmbedDocument(ctx, docID, chunksDir)
		if err != nil {
			service.logger.Errorf("EmbedDocument failed for %s: %s", docID, err)
			return nil, nil, nil, err
		}
	}

	err := db.WithTx(service.db, ctx, func(tx *sql.Tx) error {
		repo := service.repo.WithTx(tx)

		// 1. Create the conversation
		conv.Title = payload.Title
		conv.UserID = payload.UserID

		err := repo.CreateConversation(ctx, conv)
		if err != nil {
			return err
		}
		// 2. Create the first message attached to the conversation
		msg := &Message{
			ConversationID: conv.ID,
			Content:        payload.Message,
			Role:           ai.UserRole,
			DocumentIDs:    payload.DocumentIDs,
		}
		err = repo.CreateMessage(ctx, msg)
		if err != nil {
			return err
		}

		// Save message-document associations
		if err := repo.AddMessageDocuments(ctx, msg.ID, payload.DocumentIDs); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, nil, nil, err
	}

	// 4. Perform scoped semantic search
	var contextStr string
	var documentNames []string

	searchIDs := payload.DocumentIDs

	if len(searchIDs) > 0 {
		// Fetch document names for prompt
		for _, docID := range searchIDs {
			doc, err := service.docService.GetDocument(ctx, docID)
			if err != nil {
				service.logger.Errorf("GetDocument failed for %s: %s", docID, err)
				continue
			}
			documentNames = append(documentNames, doc.OriginalName)
		}

		chunks, err := service.docService.Search(ctx, payload.Message, searchIDs, 5)
		if err != nil {
			service.logger.Errorf("Search failed: %s", err)
		}
		contextStr = strings.Join(chunks, "\n\n")
	}

	// 5. Render the prompt
	chatPromptData := &ChatPromptData{
		Messages:      []Message{},
		UserMessage:   payload.Message,
		Context:       contextStr,
		DocumentNames: documentNames,
	}
	chatPrompt, err := ai.RenderPrompt(ai.ChatPrompt, chatPromptData)
	if err != nil {
		return nil, nil, nil, err
	}

	// 6. send the message to LLM
	tokenChan, errChan := service.aiService.Generate(ctx, chatPrompt)

	// 7. collect the strear tokens
	replyChan, tokenStream, errStream := service.aiService.CollectTokens(tokenChan, errChan)

	// 8. save the response to the DB
	go service.saveReply(context.Background(), conv.ID, replyChan)

	// 9. stream the response
	return conv, tokenStream, errStream, nil
}

// generate + update the conversation title
func (service *Service) handleTitleGeneration(convID string, msg string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	title, err := service.generateTitle(ctx, msg)
	if err != nil {
		service.logger.Errorf("generateTitle failed: %s using fallback", err)
		title = utils.TruncateStr(msg, 50)
	}
	updateConversationPayload := &UpdateConversationPayload{
		ConversationID: convID,
		Title:          title,
	}
	if err := service.UpdateConversation(ctx, updateConversationPayload); err != nil {
		service.logger.Errorf("updateConversationTitle failed: %s", err)
	}
}
func (service *Service) generateTitle(ctx context.Context, msg string) (string, error) {
	// 1. Render the title prompt
	promptData := &TitlePromptData{
		Message: msg,
	}
	prompt, err := ai.RenderPrompt(ai.TitlePrompt, promptData)
	if err != nil {
		return "", err
	}
	// 2. send the prompt to the LLM
	tokenChan, errChan := service.aiService.Generate(ctx, prompt)
	service.logger.Info("Title result : ", <-tokenChan, "Error: ", <-errChan)

	// 3. collect the response tokens
	replyChan, _, _ := service.aiService.CollectTokens(tokenChan, errChan)
	title, ok := <-replyChan
	if !ok || strings.TrimSpace(title) == "" {
		return "", fmt.Errorf("empty title response")
	}
	return title, nil
}
func (service *Service) UpdateConversation(ctx context.Context, payload *UpdateConversationPayload) error {
	// Validate request body
	if err := validator.Validate.Struct(payload); err != nil {
		return apierror.ErrBadRequest
	}

	conv := &Conversation{
		ID:    payload.ConversationID,
		Title: payload.Title,
	}
	return service.repo.UpdateConversation(ctx, conv)
}

// saves the AI reply when it arrives from the reply channel
func (service *Service) saveReply(
	ctx context.Context,
	conversationID string,
	replyChan <-chan string,
) {
	reply, ok := <-replyChan
	if !ok || reply == "" {
		return
	}
	msg := &Message{
		ConversationID: conversationID,
		Content:        reply,
		Role:           ai.AIRole,
	}
	err := service.repo.CreateMessage(ctx, msg)

	if err != nil {
		// ! Danger
		// TODO: Background jobs errors may Panic the server
		service.logger.Errorf("Failed to save the reply: %s", err)
	}
}

func (service *Service) CreateMessage(ctx context.Context, payload *CreateMessagePayload) (*Message, error) {
	msg := &Message{
		ConversationID: payload.ConversationID,
		Content:        payload.Content,
		Role:           payload.Role,
	}
	err := service.repo.CreateMessage(ctx, msg)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (service *Service) GetConversations(ctx context.Context, userID string) ([]*Conversation, error) {
	return service.repo.GetConversationsByUserID(ctx, userID)
}

func (service *Service) DeleteConversation(ctx context.Context, id string) error {
	return service.repo.DeleteConversation(ctx, id)
}

func (service *Service) GetConversation(ctx context.Context, id string) (*Conversation, error) {
	conv, err := service.repo.GetConversationByID(ctx, id)
	if err != nil {
		return nil, err
	}

	messages, err := service.repo.GetMessagesByConversationID(ctx, id)
	if err != nil {
		return nil, err
	}

	conv.Messages = messages
	return conv, nil
}
func (service *Service) SendMessage(ctx context.Context, payload SendMessagePayload, chunksDir string) (<-chan string, <-chan error, error) {
	// 1. Validate request body
	if err := validator.Validate.Struct(payload); err != nil {
		return nil, nil, apierror.ErrBadRequest
	}

	// 2. Check if the conversation exists
	conv, err := service.repo.GetConversationByID(ctx, payload.ConversationID)
	if err != nil {
		return nil, nil, err
	}

	// 3. Check if the conversation belongs to the user
	if conv.UserID != payload.UserID {
		return nil, nil, apierror.ErrUnauthorized
	}

	// 4. Embed all provided documents
	for _, docID := range payload.DocumentIDs {
		err := service.docService.EmbedDocument(ctx, docID, chunksDir)
		if err != nil {
			service.logger.Errorf("EmbedDocument failed for %s: %s", docID, err)
			return nil, nil, err
		}
	}

	// 5. Create the user message
	userMsg := &Message{
		ConversationID: payload.ConversationID,
		Content:        payload.Message,
		Role:           ai.UserRole,
		DocumentIDs:    payload.DocumentIDs,
	}

	err = db.WithTx(service.db, ctx, func(tx *sql.Tx) error {
		repo := service.repo.WithTx(tx)

		if err := repo.CreateMessage(ctx, userMsg); err != nil {
			return err
		}

		if err := repo.AddMessageDocuments(ctx, userMsg.ID, payload.DocumentIDs); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	// 6. Fetch previous messages
	prevMessages, err := service.repo.GetMessagesByConversationID(ctx, payload.ConversationID)
	if err != nil {
		return nil, nil, err
	}

	// 7. Perform scoped semantic search
	var contextStr string
	var documentNames []string

	// Check if its a Message scope
	isMessageScoped := service.isMessageScopedQuery(payload.Message)
	var searchIDs []string

	if isMessageScoped && len(payload.DocumentIDs) > 0 {
		searchIDs = payload.DocumentIDs
	} else {
		// All documents related to this conversation
		searchIDs, err = service.repo.GetAssociatedDocumentIDs(ctx, payload.ConversationID)
		if err != nil {
			service.logger.Errorf("GetAssociatedDocumentIDs failed: %s", err)
		}
	}

	if len(searchIDs) > 0 {
		// Fetch document names
		for _, docID := range searchIDs {
			doc, err := service.docService.GetDocument(ctx, docID)
			if err != nil {
				service.logger.Errorf("GetDocument failed for %s: %s", docID, err)
				continue
			}
			documentNames = append(documentNames, doc.OriginalName)
		}

		chunks, err := service.docService.Search(ctx, payload.Message, searchIDs, 5)
		if err != nil {
			service.logger.Errorf("Search failed: %s", err)
		}
		contextStr = strings.Join(chunks, "\n\n")
	}

	// 8. Render the prompt with history (exclude the last user message)
	history := make([]Message, 0, len(prevMessages))
	for _, m := range prevMessages {
		if m.ID == userMsg.ID {
			continue
		}
		history = append(history, m)
	}

	chatPromptData := &ChatPromptData{
		Messages:      history,
		UserMessage:   payload.Message,
		Context:       contextStr,
		DocumentNames: documentNames,
	}
	chatPrompt, err := ai.RenderPrompt(ai.ChatPrompt, chatPromptData)
	if err != nil {
		return nil, nil, err
	}

	// 9. Send to AI model and start streaming
	tokenChan, errChan := service.aiService.Generate(ctx, chatPrompt)
	replyChan, tokenStream, errStream := service.aiService.CollectTokens(tokenChan, errChan)

	// 10. Save the AI reply
	go service.saveReply(context.Background(), payload.ConversationID, replyChan)

	return tokenStream, errStream, nil
}

// isMessageScopedQuery checks if the message is scoped to this message files or for the chat files
func (service *Service) isMessageScopedQuery(query string) bool {
	pattern := `(?i)\b(this|these|those|that|current|the)\s+(?:current\s+)?(file|document|doc)s?\b`
	matched, _ := regexp.MatchString(pattern, query)
	return matched
}
