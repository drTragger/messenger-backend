package requests

type CreateChatRequest struct {
	UserID uint `json:"userId" validate:"required,gt=0"`
}
