package ctrl

import "errors"

// ErrNotFound is returned when a resource is not found.
var ErrNotFound = errors.New("not found")

// ErrAlreadyExists is returned when a resource already exists.
var ErrAlreadyExists = errors.New("already exists")

// ErrParseUUID is returned when uuid is not valid.
var ErrParseUUID = errors.New("failed to parse uuid")

// ErrCodeIsNotValid is returned when login code is not valid.
var ErrCodeIsNotValid = errors.New("code is not valid")

// ErrWhileGeneratingToken is returned when error while generating token.
var ErrWhileGeneratingToken = errors.New("error while generating token")
