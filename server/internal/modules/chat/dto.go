package chat

type StartConversationDTO struct {
	Message     string   `json:"message" validate:"required"`
	DocumentIDs []string `json:"document_ids" validate:"omitempty,dive,uuid"`
}

type UpdateConversationDTO struct {
	Title string `json:"title" validate:"required"`
}

type SendMessageDTO struct {
	Message     string   `json:"message" validate:"required"`
	DocumentIDs []string `json:"document_ids" validate:"omitempty,dive,uuid"`
}
