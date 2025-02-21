package validation

import "errors"

var ErrLoginIsRequired = errors.New("login is required")
var ErrMissingToken = errors.New("token is required")
var ErrMissingCode = errors.New("code is required")
var ErrMissingEmail = errors.New("email is required")
var ErrInvalidEmail = errors.New("invalid email format")
var ErrMissingPass = errors.New("missing password")
var ErrPassTooShort = errors.New("password must be at least 8 characters long")
var ErrMissingName = errors.New("missing name")
