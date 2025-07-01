package providers

import "errors"

var (
	// ErrUnknownProvider is error that indicates unknown provider.
	ErrUnknownProvider = errors.New("unknown provider")

	// ErrInvalidStateFormat is error that indicates invalid state format.
	ErrInvalidStateFormat = errors.New("invalid oauth state format")

	// ErrInvalidSignature is error that indicates invalid signature.
	ErrInvalidSignature = errors.New("invalid signature")

	// ErrInvalidDataFormat is error that indicates invalid data format.
	ErrInvalidDataFormat = errors.New("invalid data format")

	// ErrStateExpired is error that indicates state expired.
	ErrStateExpired = errors.New("state expired")
)
