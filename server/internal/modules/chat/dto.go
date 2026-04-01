package chat

type StartConversationDTO struct {
	Message string `json:"message" validate:"required"`
}

type UpdateConversationDTO struct {
	Title string `json:"title" validate:"required"`
}

type SendMessageDTO struct {
	Message string `json:"message" validate:"required"`
}
