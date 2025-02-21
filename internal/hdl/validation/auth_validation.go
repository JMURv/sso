package validation

import (
	"github.com/JMURv/sso/internal/dto"
)

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
