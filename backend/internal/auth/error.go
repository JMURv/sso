package auth

import "errors"

var (
	// ErrInvalidCredentials is error that indicates invalid credentials.
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrTokenRevoked is error that indicates token expired.
	ErrTokenRevoked = errors.New("token revoked")
)
