package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

type Code string

const (
	CodeValidation   Code = "VALIDATION_ERROR"
	CodeUnauthorized Code = "UNAUTHORIZED"
	CodeForbidden    Code = "FORBIDDEN"
	CodeNotFound     Code = "NOT_FOUND"
	CodeConflict     Code = "CONFLICT"
	CodeBadRequest   Code = "BAD_REQUEST"
	CodeRateLimited  Code = "RATE_LIMITED"
	CodeInternal     Code = "INTERNAL_ERROR"
)

type AppError struct {
	Status  int
	Code    Code
	Message string
	Fields  map[string]string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.Err }

func New(status int, code Code, message string) *AppError {
	return &AppError{Status: status, Code: code, Message: message}
}

func Wrap(status int, code Code, message string, err error) *AppError {
	return &AppError{Status: status, Code: code, Message: message, Err: err}
}

func Validation(fields map[string]string) *AppError {
	return &AppError{
		Status:  http.StatusUnprocessableEntity,
		Code:    CodeValidation,
		Message: "Validation failed",
		Fields:  fields,
	}
}

func NotFound(message string) *AppError {
	return New(http.StatusNotFound, CodeNotFound, message)
}

func Unauthorized(message string) *AppError {
	return New(http.StatusUnauthorized, CodeUnauthorized, message)
}

func Forbidden(message string) *AppError {
	return New(http.StatusForbidden, CodeForbidden, message)
}

func Conflict(message string) *AppError {
	return New(http.StatusConflict, CodeConflict, message)
}

func BadRequest(message string) *AppError {
	return New(http.StatusBadRequest, CodeBadRequest, message)
}

func Internal(err error) *AppError {
	return Wrap(http.StatusInternalServerError, CodeInternal, "Internal server error", err)
}

func As(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
