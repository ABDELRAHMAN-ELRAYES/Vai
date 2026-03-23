package chat

type StartConversationPayload struct {
	UserID  string
	Title   string
	Message string
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
	Messages    []string
	UserMessage string
}

type TitlePromptData struct {
	Message string
}
