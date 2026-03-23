package chat

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/shared"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/users"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/validator"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/apierror"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/httputil"
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
		UserID:  user.ID,
		Title:   "Default",
		Message: startConversationDTO.Message,
	}
	conversation, responseStream, errStream, err := handler.service.StartConversation(ctx, *startConversationPayload)
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
	convIDEvent := fmt.Sprintf("data: {\"conversation_id\": \"%s\"}\n\n", conversation.ID)
	_, _ = w.Write([]byte(convIDEvent))
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
			_, _ = w.Write([]byte("data: " + token + "\n\n"))
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
