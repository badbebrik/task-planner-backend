package errors

import "errors"

var (
	ErrInternalServer  = errors.New("internal server error")
	ErrBadRequest      = errors.New("bad request")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrForbidden       = errors.New("forbidden")
	ErrNotFound        = errors.New("resource not found")
	ErrTooManyRequests = errors.New("too many requests")
)
