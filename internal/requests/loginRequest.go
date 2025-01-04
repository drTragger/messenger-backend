package requests

// LoginRequest defines the payload for the login endpoint
type LoginRequest struct {
	Phone    string `json:"phone" validate:"required,phone"`
	Password string `json:"password" validate:"required"`
}
