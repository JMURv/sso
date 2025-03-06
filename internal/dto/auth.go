package dto

import "github.com/google/uuid"

type DeviceRequest struct {
	IP string `json:"ip"`
	UA string `json:"ua"`
}

type GenPairResponse struct {
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
}

type RefreshRequest struct {
	Refresh string `json:"refresh"`
}

type RefreshResponse struct {
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
}

type EmailAndPasswordRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type EmailAndPasswordResponse struct {
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
}

type LoginCodeRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CheckLoginCodeRequest struct {
	Email string `json:"email"`
	Code  int    `json:"code"`
}

type CheckLoginCodeResponse struct {
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
}

type CheckEmailRequest struct {
	Email string `json:"email"`
}

type CheckEmailResponse struct {
	Exists bool `json:"exists"`
}

type CheckForgotPasswordEmailRequest struct {
	Password string    `json:"password"`
	ID       uuid.UUID `json:"uidb64"`
	Code     int       `json:"token"`
}

type SendForgotPasswordEmail struct {
	Email string `json:"email"`
}

type SendSupportEmailRequest struct {
	Theme string `json:"theme"`
	Text  string `json:"text"`
}

type TokenRequest struct {
	Token string `json:"token"`
}
