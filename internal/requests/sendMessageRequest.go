package requests

// SendMessageRequest defines the payload for the send message endpoint
type SendMessageRequest struct {
	RecipientID uint    `json:"recipientId" validate:"required,gt=0"`
	ParentID    *uint   `json:"parentId" validate:"omitempty,gt=0"`
	Content     *string `json:"content" validate:"omitempty,min=1,max=5000"`
}
