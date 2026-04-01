package chat

type StartConversationPayload struct {
	UserID  string `json:"user_id" validate:"required"`
	Title   string `json:"title" validate:"required"`
	Message string `json:"message" validate:"required"`
}

type SendMessagePayload struct {
	ConversationID string `json:"conversation_id" validate:"required"`
	UserID         string `json:"user_id" validate:"required"`
	Message        string `json:"message" validate:"required"`
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
	Messages    []Message
	UserMessage string
}

type TitlePromptData struct {
	Message string
}
