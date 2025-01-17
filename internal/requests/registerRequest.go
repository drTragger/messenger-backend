package requests

// RegisterRequest defines the payload for the register endpoint
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50,alphaunicode"` // Supports Unicode
	Phone    string `json:"phone" validate:"required,phone"`
	Password string `json:"password" validate:"required,min=6,max=50"`
}
