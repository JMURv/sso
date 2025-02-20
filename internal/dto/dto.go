package dto

type EmailAndPasswordRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type EmailAndPasswordResponse struct {
	Token string `json:"token"`
}
