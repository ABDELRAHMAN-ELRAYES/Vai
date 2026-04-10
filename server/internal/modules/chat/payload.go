package chat

type StartConversationPayload struct {
	UserID      string   `json:"user_id" validate:"required"`
	DocumentIDs []string `json:"document_ids" validate:"omitempty"`
	Title       string   `json:"title" validate:"required"`
	Message     string   `json:"message" validate:"required"`
}

type SendMessagePayload struct {
	ConversationID string   `json:"conversation_id" validate:"required"`
	UserID         string   `json:"user_id" validate:"required"`
	DocumentIDs    []string `json:"document_ids" validate:"omitempty"`
	Message        string   `json:"message" validate:"required"`
}

type CreateMessagePayload struct {
	ConversationID string
	Content        string
	Role           string
}

type UpdateConversationPayload struct {
	ConversationID string
	Title          string
}
type ChatPromptData struct {
	Messages      []Message
	UserMessage   string
	Context       string
	DocumentNames []string
}

type TitlePromptData struct {
	Message string
}
