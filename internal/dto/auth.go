package dto

import "github.com/google/uuid"

type EmailAndPasswordRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type EmailAndPasswordResponse struct {
	Token string `json:"token"`
}

type LoginCodeRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CheckLoginCodeRequest struct {
	Email string `json:"email"`
	Code  int    `json:"code"`
}

type CheckEmailRequest struct {
	Email string `json:"email"`
}

type CheckEmailResponse struct {
	Exists bool `json:"exists"`
}

type CheckForgotPasswordEmailRequest struct {
	Password string    `json:"password"`
	Uidb64   uuid.UUID `json:"uidb64"`
	Token    int       `json:"token"`
}

type SendSupportEmailRequest struct {
	Theme string `json:"theme"`
	Text  string `json:"text"`
}

type TokenRequest struct {
	Token string `json:"token"`
}
