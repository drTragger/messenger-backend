package requests

// VerifyCodeRequest defines the payload for the verify phone code endpoint
type VerifyCodeRequest struct {
	Phone string `json:"phone" validate:"required,phone"`
	Code  string `json:"code" validate:"required,len=6"`
}
