package requests

type ResendCodeRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Phone    string `json:"phone" validate:"required,phone"`
}
