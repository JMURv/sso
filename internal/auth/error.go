package auth

import "errors"

var ErrFailedToParseClaims = errors.New("failed to parse claims")
var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrInvalidToken = errors.New("invalid token")
var ErrWhileCreatingToken = errors.New("error while creating token")
var ErrUnexpectedSignMethod = errors.New("unexpected signing method")
var ErrTokenRevoked = errors.New("token revoked")
