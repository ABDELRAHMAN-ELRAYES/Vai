package chat

type StartConversationDTO struct {
	Message    string `json:"message" validate:"required"`
	DocumentID string `json:"document_id" validate:"omitempty,uuid"`
}

type UpdateConversationDTO struct {
	Title string `json:"title" validate:"required"`
}

type SendMessageDTO struct {
	Message    string `json:"message" validate:"required"`
	DocumentID string `json:"document_id" validate:"omitempty,uuid"`
}
