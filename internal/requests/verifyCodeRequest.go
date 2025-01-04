package requests

type VerifyCodeRequest struct {
	Phone string `json:"phone" validate:"required,phone"`
	Code  string `json:"code" validate:"required,len=6"`
}
