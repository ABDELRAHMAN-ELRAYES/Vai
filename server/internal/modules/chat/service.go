package chat

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/db"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/ai"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/users"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/validator"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/apierror"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/utils"
	"go.uber.org/zap"
)

type Service struct {
	db          *sql.DB
	repo        *Repository
	aiService   *ai.Service
	logger      *zap.SugaredLogger
	userService *users.Service
}

func NewService(db *sql.DB, repo *Repository, aiService *ai.Service, logger *zap.SugaredLogger, userService *users.Service) *Service {
	return &Service{
		db:          db,
		repo:        repo,
		aiService:   aiService,
		logger:      logger,
		userService: userService,
	}
}
func (service *Service) StartConversation(ctx context.Context, payload StartConversationPayload) (*Conversation, <-chan string, <-chan error, error) {
	// Validate request body
	if err := validator.Validate.Struct(payload); err != nil {
		return nil, nil, nil, apierror.ErrBadRequest
	}

	conv := &Conversation{}

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
		}
		err = repo.CreateMessage(ctx, msg)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, nil, nil, err
	}
	// 3. Generate + Update a new conversation title
	// Accordin to the first submitted user message
	// ! Error, Context Done
	// go service.handleTitleGeneration(conv.ID, payload.Message)

	// 4. Render the prompt
	chatPromptData := &ChatPromptData{
		Messages:    []Message{},
		UserMessage: payload.Message,
	}
	chatPrompt, err := ai.RenderPrompt(ai.ChatPrompt, chatPromptData)
	if err != nil {
		return nil, nil, nil, err
	}
	// 5. send the message to LLM
	tokenChan, errChan := service.aiService.Generate(ctx, chatPrompt)

	// 6. collect the strear tokens
	replyChan, tokenStream, errStream := service.aiService.CollectTokens(tokenChan, errChan)

	// 7. save the response to the DB
	go service.saveReply(context.Background(), conv.ID, replyChan)

	// 8. stream the response
	return conv, tokenStream, errStream, nil
}

// generate + update the conversation title
func (service *Service) handleTitleGeneration(convID string, msg string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	title, err := service.generateTitle(ctx, msg)
	if err != nil {
		service.logger.Errorf("generateTitle failed: %s — using fallback", err)
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
	service.logger.Info("Title result : ",<-tokenChan,"Error: ",<-errChan)

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
func (service *Service) SendMessage(ctx context.Context, payload SendMessagePayload) (<-chan string, <-chan error, error) {
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

	// 4. Create the user message
	userMsg := &Message{
		ConversationID: payload.ConversationID,
		Content:        payload.Message,
		Role:           ai.UserRole,
	}
	err = service.repo.CreateMessage(ctx, userMsg)
	if err != nil {
		return nil, nil, err
	}

	// 4. Fetch previous messages
	prevMessages, err := service.repo.GetMessagesByConversationID(ctx, payload.ConversationID)
	if err != nil {
		return nil, nil, err
	}

	// 5. Render the prompt with history (exclude the last user message)
	history := make([]Message, 0, len(prevMessages))
	for _, m := range prevMessages {
		if m.ID == userMsg.ID {
			continue
		}
		history = append(history, m)
	}

	chatPromptData := &ChatPromptData{
		Messages:    history,
		UserMessage: payload.Message,
	}
	chatPrompt, err := ai.RenderPrompt(ai.ChatPrompt, chatPromptData)
	if err != nil {
		return nil, nil, err
	}

	// 6. Send to AI model and start streaming
	tokenChan, errChan := service.aiService.Generate(ctx, chatPrompt)
	replyChan, tokenStream, errStream := service.aiService.CollectTokens(tokenChan, errChan)

	// 7. Save the AI reply 
	go service.saveReply(context.Background(), payload.ConversationID, replyChan)

	return tokenStream, errStream, nil
}
