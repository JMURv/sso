package handler

import "errors"

var ErrRetrievePathVars = errors.New("failed to retrieve path variable")
var ErrMethodNotAllowed = errors.New("method not allowed")
var ErrMissingEmail = errors.New("email is required")
var ErrEmailAndPasswordRequired = errors.New("email and password are required")
var ErrEmailAndCodeRequired = errors.New("email and code are required")
