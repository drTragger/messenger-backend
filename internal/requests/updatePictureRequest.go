package requests

type ProfilePictureRequest struct {
	Picture string `validate:"required,oneof=image/jpeg image/png"`
}
