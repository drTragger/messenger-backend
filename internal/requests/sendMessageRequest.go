package requests

// SendMessageRequest defines the payload for the send message endpoint
type SendMessageRequest struct {
	RecipientID uint   `json:"recipientId" validate:"required,gt=0"`
	ChatID      uint   `json:"chatId" validate:"omitempty,gt=0"`
	Message     string `json:"message" validate:"required,min=1,max=5000"`
}
