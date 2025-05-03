package auth

import "errors"

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrTokenRevoked = errors.New("token revoked")
