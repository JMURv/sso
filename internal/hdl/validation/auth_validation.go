package validation

import (
	"github.com/JMURv/sso/internal/dto"
	"github.com/google/uuid"
)

func RefreshRequest(req *dto.RefreshRequest) error {
	if req.Refresh == "" {
		return ErrMissingToken
	}

	return nil
}

func LoginAndPasswordRequest(req *dto.EmailAndPasswordRequest) error {
	if req.Email == "" {
		return ErrMissingEmail
	}

	if req.Password == "" {
		return ErrMissingPass
	}

	return nil
}

func LoginCodeRequest(req *dto.LoginCodeRequest) error {
	if req.Email == "" {
		return ErrMissingEmail
	}

	if req.Password == "" {
		return ErrMissingPass
	}

	return ValidateEmail(req.Email)
}

func CheckLoginCodeRequest(req *dto.CheckLoginCodeRequest) error {
	if req.Email == "" {
		return ErrMissingEmail
	}

	if req.Code == 0 {
		return ErrMissingCode
	}

	return ValidateEmail(req.Email)
}

func TokenRequest(req *dto.TokenRequest) error {
	if req.Token == "" {
		return ErrMissingToken
	}
	return nil
}

func SendForgotPasswordEmail(req *dto.SendForgotPasswordEmail) error {
	if req.Email == "" {
		return ErrMissingToken
	}

	return ValidateEmail(req.Email)
}

func CheckForgotPasswordEmailRequest(req *dto.CheckForgotPasswordEmailRequest) error {
	if req.ID == uuid.New() {
		return ErrIDIsRequired
	}

	if req.Password == "" {
		return ErrMissingPass
	}

	if req.Code == 0 {
		return ErrCodeIsRequired
	}

	return nil
}
