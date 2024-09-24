package handler

import "errors"

var ErrMissingEmail = errors.New("email is required")
var ErrEmailAndPasswordRequired = errors.New("email and password are required")
var ErrEmailAndCodeRequired = errors.New("email and code are required")
