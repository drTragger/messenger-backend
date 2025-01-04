package responses

type LoginResponse struct {
	Token  string `json:"token"`
	Expire int64  `json:"expire"`
}
