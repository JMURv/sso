package ctrl

import "errors"

var ErrNotFound = errors.New("not found")

var ErrAlreadyExists = errors.New("already exists")
var ErrInternalError = errors.New("internal error")
var ErrDecodeRequest = errors.New("failed to decode request")
var ErrUnauthorized = errors.New("unauthorized")
var ErrParseUUID = errors.New("failed to parse uuid")
var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrCodeIsNotValid = errors.New("code is not valid")
var ErrWhileGeneratingToken = errors.New("error while generating token")
