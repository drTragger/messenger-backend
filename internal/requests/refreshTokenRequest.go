package requests

type RefreshTokenRequest struct {
	Token string `json:"token" validate:"required"`
}
