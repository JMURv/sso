package dto

type LoginStartRequest struct {
	Email string `json:"email" validate:"required,email"`
}
