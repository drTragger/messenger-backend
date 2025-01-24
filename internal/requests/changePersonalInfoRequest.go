package requests

type ChangePersonalInfoRequest struct {
	FirstName string  `json:"firstName" validate:"required,min=2,max=50"`
	LastName  *string `json:"lastName" validate:"omitempty,max=80"`
}
