package requests

type SendMessageRequest struct {
	RecipientID uint   `json:"recipientId"`
	Message     string `json:"message"`
}
