package repository

import "errors"

var ErrNotFound = errors.New("not found")
var ErrAlreadyExists = errors.New("already exists")
var ErrGeneratingPassword = errors.New("generating password")
