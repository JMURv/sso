package validation

import "github.com/JMURv/sso/internal/dto"

func LoginAndPasswordRequest(req *dto.EmailAndPasswordRequest) error {
	if req.Email == "" {
		return ErrMissingEmail
	}

	if req.Password == "" {
		return ErrMissingPass
	}

	return nil
}
