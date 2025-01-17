package requests

type EditMessageRequest struct {
	Content string `json:"content" validate:"required,min=1,max=5000"`
}
