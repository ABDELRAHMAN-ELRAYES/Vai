package chat

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/shared"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/users"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/validator"
	apierror "github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/errors"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/httputil"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	app     *app.Application
	service *Service
}

func NewHandler(app *app.Application, service *Service) *Handler {
	return &Handler{
		service: service,
		app:     app,
	}
}

// StartConversation godoc
//
//	@Summary		Start a new conversation
//	@Description	Creates a new conversation, sends the first message to the LLM, and streams the response back using Server-Sent Events (SSE). The first SSE event contains the conversation_id, subsequent events contain tokens, and the final event is [DONE].
//	@Tags			conversations
//	@Accept			json
//	@Produce		text/event-stream
//	@Param			payload	body		StartConversationDTO	true	"First message payload"
//	@Success		200		{string}	string					"SSE stream — first event: {conversation_id}, then tokens, then [DONE]"
//	@Failure		400		{object}	error					"Invalid request body"
//	@Failure		401		{object}	error					"Unauthorized"
//	@Failure		500		{object}	error					"Internal server error"
//	@Security		BearerAuth
//	@Router			/conversations [post]
func (handler *Handler) StartConversation(w http.ResponseWriter, r *http.Request) {
	logger := handler.app.Logger
	// 1. Read the request body
	var startConversationDTO StartConversationDTO
	if err := httputil.ReadJSON(w, r, &startConversationDTO); err != nil {
		apierror.BadRequest(logger, w, r, err)
		return
	}
	// 2. Validate the request body
	if err := validator.Validate.Struct(startConversationDTO); err != nil {
		apierror.BadRequest(logger, w, r, err)
		return
	}

	// 3. Get user from the request context
	ctx := r.Context()

	user := ctx.Value(shared.UserCtxKey).(*users.User)
	// 4. Form the service payload
	startConversationPayload := &StartConversationPayload{
		UserID:      user.ID,
		Title:       "Default",
		Message:     startConversationDTO.Message,
		DocumentIDs: startConversationDTO.DocumentIDs,
	}
	conversation, responseStream, errStream, err := handler.service.StartConversation(ctx, *startConversationPayload, handler.app.Config.RAG.Chunker.ChunksDir)
	if err != nil {
		switch err {
		case apierror.ErrNotFound:
			apierror.NotFound(logger, w, r, err)
			return
		default:
			apierror.InternalServerError(logger, w, r, err)
			return
		}
	}
	// 5. Stream the response
	// 5.1 Setup the SSE required header
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// 5.2 Ensure to send each token immediately
	flusher, ok := w.(http.Flusher)
	if !ok {
		apierror.InternalServerError(logger, w, r, errors.New("streaming unsupported"))
		return
	}

	// 5.3 Send the conversation ID as the first Event
	infoData := map[string]string{
		"conversationId": conversation.ID,
		"type":           "info",
	}
	infoJSON, _ := json.Marshal(infoData)
	_, _ = w.Write([]byte(fmt.Sprintf("data: %s\n\n", infoJSON)))
	flusher.Flush()

	// 5.4 Stream the reply
	for {
		select {

		case token, ok := <-responseStream:
			if !ok {
				_, _ = w.Write([]byte("data: [DONE]\n\n"))
				flusher.Flush()
				return
			}
			msgData := map[string]string{
				"token": token,
				"type":  "token",
			}
			jsonData, _ := json.Marshal(msgData)
			_, _ = w.Write([]byte(fmt.Sprintf("data: %s\n\n", jsonData)))
			flusher.Flush()

		case err, ok := <-errStream:
			if !ok {
				continue
			}
			if err != nil {
				apierror.InternalServerError(logger, w, r, err)
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

// GetConversations godoc
//
//	@Summary		Get all user conversations
//	@Description	Returns a list of all conversations for the authenticated user, ordered by last update time.
//	@Tags			conversations
//	@Produce		json
//	@Success		200		{array}		Conversation			"List of conversations"
//	@Failure		401		{object}	error					"Unauthorized"
//	@Failure		500		{object}	error					"Internal server error"
//	@Security		BearerAuth
//	@Router			/conversations [get]
func (handler *Handler) GetConversations(w http.ResponseWriter, r *http.Request) {
	logger := handler.app.Logger
	ctx := r.Context()
	user := ctx.Value(shared.UserCtxKey).(*users.User)

	conversations, err := handler.service.GetConversations(ctx, user.ID)
	if err != nil {
		apierror.InternalServerError(logger, w, r, err)
		return
	}

	if err := httputil.WriteJSON(w, http.StatusOK, conversations); err != nil {
		apierror.InternalServerError(logger, w, r, err)
		return
	}
}

// UpdateConversation godoc
//
//	@Summary		Update conversation title
//	@Description	Updates the title of an existing conversation.
//	@Tags			conversations
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"Conversation ID"
//	@Param			payload	body		UpdateConversationDTO	true	"New title payload"
//	@Success		204		{string}	string					"No Content"
//	@Failure		400		{object}	error					"Invalid request body"
//	@Failure		401		{object}	error					"Unauthorized"
//	@Failure		404		{object}	error					"Conversation not found"
//	@Failure		500		{object}	error					"Internal server error"
//	@Security		BearerAuth
//	@Router			/conversations/{id} [patch]
func (handler *Handler) UpdateConversation(w http.ResponseWriter, r *http.Request) {
	logger := handler.app.Logger
	id := chi.URLParam(r, "id")

	var updateConversationDto UpdateConversationDTO
	if err := httputil.ReadJSON(w, r, &updateConversationDto); err != nil {
		apierror.BadRequest(logger, w, r, err)
		return
	}

	if err := validator.Validate.Struct(updateConversationDto); err != nil {
		apierror.BadRequest(logger, w, r, err)
		return
	}

	payload := &UpdateConversationPayload{
		ConversationID: id,
		Title:          updateConversationDto.Title,
	}

	if err := handler.service.UpdateConversation(r.Context(), payload); err != nil {
		if err == apierror.ErrNotFound {
			apierror.NotFound(logger, w, r, err)
			return
		}
		apierror.InternalServerError(logger, w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteConversation godoc
//
//	@Summary		Delete a conversation
//	@Description	Deletes an existing conversation and all its messages.
//	@Tags			conversations
//	@Param			id		path		string					true	"Conversation ID"
//	@Success		204		{string}	string					"No Content"
//	@Failure		401		{object}	error					"Unauthorized"
//	@Failure		404		{object}	error					"Conversation not found"
//	@Failure		500		{object}	error					"Internal server error"
//	@Security		BearerAuth
//	@Router			/conversations/{id} [delete]
func (handler *Handler) DeleteConversation(w http.ResponseWriter, r *http.Request) {
	logger := handler.app.Logger
	id := chi.URLParam(r, "id")

	if err := handler.service.DeleteConversation(r.Context(), id); err != nil {
		if err == apierror.ErrNotFound {
			apierror.NotFound(logger, w, r, err)
			return
		}
		apierror.InternalServerError(logger, w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetChat godoc
//
//	@Summary		Get a conversation
//	@Description	Returns a single conversation and its messages.
//	@Tags			conversations
//	@Param			id	path		string			true	"Conversation ID"
//	@Produce		json
//	@Success		200	{object}	Conversation	"Conversation object"
//	@Failure		401	{object}	error			"Unauthorized"
//	@Failure		404	{object}	error			"Conversation not found"
//	@Failure		500	{object}	error			"Internal server error"
//	@Security		BearerAuth
//	@Router			/conversations/{id} [get]
func (handler *Handler) GetChat(w http.ResponseWriter, r *http.Request) {
	logger := handler.app.Logger
	id := chi.URLParam(r, "id")

	conversation, err := handler.service.GetConversation(r.Context(), id)
	if err != nil {
		if err == apierror.ErrNotFound {
			apierror.NotFound(logger, w, r, err)
			return
		}
		apierror.InternalServerError(logger, w, r, err)
		return
	}

	if err := httputil.WriteJSON(w, http.StatusOK, conversation); err != nil {
		apierror.InternalServerError(logger, w, r, err)
		return
	}
}

// SendMessage godoc
//
//	@Summary		Continue conversation
//	@Description	Adds a new message to an existing conversation, sends it to the LLM, and streams the response back using Server-Sent Events (SSE).
//	@Tags			conversations
//	@Accept			json
//	@Produce		text/event-stream
//	@Param			id		path		string			true	"Conversation ID"
//	@Param			payload	body		SendMessageDTO	true	"Message payload"
//	@Success		200		{string}	string			"SSE stream tokens"
//	@Failure		400		{object}	error			"Invalid request"
//	@Failure		401		{object}	error			"Unauthorized"
//	@Failure		404		{object}	error			"Conversation not found"
//	@Failure		500		{object}	error			"Internal server error"
//	@Security		BearerAuth
//	@Router			/conversations/{id} [post]
func (handler *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	logger := handler.app.Logger
	id := chi.URLParam(r, "id")

	// 1. Read the request body
	var sendMessageDTO SendMessageDTO
	if err := httputil.ReadJSON(w, r, &sendMessageDTO); err != nil {
		apierror.BadRequest(logger, w, r, err)
		return
	}

	// 2. Validate the request body
	if err := validator.Validate.Struct(sendMessageDTO); err != nil {
		apierror.BadRequest(logger, w, r, err)
		return
	}

	// 2. Get user context
	ctx := r.Context()
	user := ctx.Value(shared.UserCtxKey).(*users.User)

	// 3. Form the service payload
	payload := SendMessagePayload{
		ConversationID: id,
		UserID:         user.ID,
		Message:        sendMessageDTO.Message,
		DocumentIDs:    sendMessageDTO.DocumentIDs,
	}

	tokenStream, errStream, err := handler.service.SendMessage(ctx, payload, handler.app.Config.RAG.Chunker.ChunksDir)
	if err != nil {
		switch err {
		case apierror.ErrNotFound:
			apierror.NotFound(logger, w, r, err)
			return
		case apierror.ErrUnauthorized:
			apierror.Unauthorized(logger, w, r, err)
			return
		default:
			apierror.InternalServerError(logger, w, r, err)
			return
		}
	}

	// 4. Setup SSE headers and stream the response
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		apierror.InternalServerError(logger, w, r, errors.New("streaming unsupported"))
		return
	}

	for {
		select {
		case token, ok := <-tokenStream:
			if !ok {
				_, _ = w.Write([]byte("data: [DONE]\n\n"))
				flusher.Flush()
				return
			}
			msgData := map[string]string{
				"token": token,
				"type":  "token",
			}
			jsonData, _ := json.Marshal(msgData)
			_, _ = w.Write([]byte(fmt.Sprintf("data: %s\n\n", jsonData)))
			flusher.Flush()

		case err, ok := <-errStream:
			if !ok {
				continue
			}
			if err != nil {
				apierror.InternalServerError(logger, w, r, err)
				return
			}

		case <-ctx.Done():
			return
		}
	}
}
