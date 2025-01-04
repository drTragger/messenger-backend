package requests

// LoginRequest defines the payload for the login endpoint
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
