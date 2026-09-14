package repository

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrEmailExists   = errors.New("email already exists")
	ErrInvalidToken  = errors.New("invalid or expired token")
	ErrTokenNotFound = errors.New("token not found")
)
