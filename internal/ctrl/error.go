package ctrl

import "errors"

var ErrNotFound = errors.New("not found")
var ErrAlreadyExists = errors.New("already exists")
var ErrParseUUID = errors.New("failed to parse uuid")
var ErrCodeIsNotValid = errors.New("code is not valid")
var ErrWhileGeneratingToken = errors.New("error while generating token")
