package chat

type StartConversationDTO struct {
	Message string `json:"message" validate:"required"`
}
