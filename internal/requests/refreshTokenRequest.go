package requests

// RefreshTokenRequest defines the payload for the refresh token endpoint
type RefreshTokenRequest struct {
	Token string `json:"token" validate:"required,jwt"`
}
