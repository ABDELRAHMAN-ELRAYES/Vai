package chat

import "github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/documents"

type Conversation struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Messages  []Message `json:"messages"`
}

type Message struct {
	ID             string               `json:"id"`
	ConversationID string               `json:"conversation_id"`
	Content        string               `json:"content"`
	Role           string               `json:"role"`
	CreatedAt      string               `json:"created_at"`
	DocumentIDs    []string             `json:"document_ids,omitempty"`
	Documents      []documents.Document `json:"documents,omitempty"`
}
