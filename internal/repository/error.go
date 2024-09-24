package repository

import "errors"

var ErrNotFound = errors.New("not found")
var ErrAlreadyExists = errors.New("already exists")
var ErrUsernameIsRequired = errors.New("username is required")
var ErrEmailIsRequired = errors.New("email is required")
var ErrPasswordIsRequired = errors.New("password is required")
var ErrGeneratingPassword = errors.New("generating password")
