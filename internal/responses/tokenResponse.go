package responses

type TokenResponse struct {
	Token   string `json:"token"`
	Expires int64  `json:"expires"`
}
