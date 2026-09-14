package handler

import (
	"net/http"

	"github.com/saurav11sarkar/go-production-boilerplate/internal/apperror"
)

func badJSON(err error) error {
	return apperror.Wrap(http.StatusBadRequest, apperror.CodeBadRequest, "Invalid request body", err)
}

func validation(fields map[string]string) error {
	return apperror.Validation(fields)
}
