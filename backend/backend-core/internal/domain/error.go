package domain

import "errors"

var ErrNotFound = errors.New("not found")
var ErrAlreadyExists = errors.New("already exists")
var ErrBarRequest = errors.New("bad request")
