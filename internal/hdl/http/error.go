package http

import "errors"

var ErrMethodNotAllowed = errors.New("method not allowed")
var ErrRetrievePathVars = errors.New("cannot retrieve path variables")
