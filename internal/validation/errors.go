package validation

import "errors"

var ErrMissingEmail = errors.New("email is required")
var ErrInvalidEmail = errors.New("invalid email format")
var ErrMissingPass = errors.New("missing password")
var ErrPassTooShort = errors.New("password must be at least 8 characters long")
var ErrMissingName = errors.New("missing name")
