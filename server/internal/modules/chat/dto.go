package chat

type CreateFirstConversationPayload struct {
	UserID  string
	Title   string
	Message string
}

type CreateMessagePayload struct {
	ConversationID string
	Content        string
	Role           string
}
