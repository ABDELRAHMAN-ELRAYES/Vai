package chat

type Conversation struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Messages  []Message `json:"messages"`
}

type Message struct {
	ID             string `json:"id"`
	ConversationID string `json:"conversation_id"`
	Content        string `json:"content"`
	Role           string `json:"role"`
	CreatedAt      string `json:"created_at"`
}
type UpdateConversationPayload struct {
	ConversationID string
	Title          string
}
type ChatPromptData struct {
	Messages    []string
	UserMessage string
}

type TitlePromptData struct {
	Message string
}
