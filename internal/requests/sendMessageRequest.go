package requests

type SendMessageRequest struct {
	RecipientID uint   `json:"recipientId" validate:"required"`
	Message     string `json:"message" validate:"required"`
}
