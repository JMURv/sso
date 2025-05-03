package providers

import "errors"

var ErrUnknownProvider = errors.New("unknown provider")
var ErrInvalidStateFormat = errors.New("invalid oauth state format")
var ErrInvalidSignature = errors.New("invalid signature")
var ErrInvalidDataFormat = errors.New("invalid data format")
var ErrStateExpired = errors.New("state expired")
