package captcha

import (
	"errors"
)

var (
	// ErrVerificationFailed is error that signals CAPTCHA verification failure.
	ErrVerificationFailed = errors.New("CAPTCHA verification failed")

	// ErrValidationFailed is error that signals CAPTCHA validation failure.
	ErrValidationFailed = errors.New("CAPTCHA validation failed")
)
