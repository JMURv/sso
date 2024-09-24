package validation

import (
	"github.com/JMURv/sso/pkg/model"
	"regexp"
)

func ValidateEmail(email string) error {
	if compile, _ := regexp.Compile(`^[\w.%+-]+@[\w.-]+\.[a-zA-Z]+$`); !compile.MatchString(email) {
		return ErrInvalidEmail
	}
	return nil
}

func NewUserValidation(u *model.User) error {
	if u.Name == "" {
		return ErrMissingName
	}
	if u.Email == "" {
		return ErrMissingEmail
	}

	if compile, _ := regexp.Compile(`^[\w.%+-]+@[\w.-]+\.[a-zA-Z]+$`); !compile.MatchString(u.Email) {
		return ErrInvalidEmail
	}

	if u.Password == "" {
		return ErrMissingPass
	}

	if len(u.Password) < 8 {
		return ErrPassTooShort
	}

	return nil
}

func UserValidation(u *model.User) error {
	if u.Name == "" {
		return ErrMissingName
	}

	if u.Email == "" {
		return ErrMissingEmail
	}

	if compile, _ := regexp.Compile(`^[\w.%+-]+@[\w.-]+\.[a-zA-Z]+$`); !compile.MatchString(u.Email) {
		return ErrInvalidEmail
	}

	return nil
}
