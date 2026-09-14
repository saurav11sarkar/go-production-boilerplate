package httpx

import (
	"log/slog"
	"net/http"

	"github.com/saurav11sarkar/go-production-boilerplate/internal/apperror"
)

type errorBody struct {
	Code    apperror.Code     `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func Error(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	if appErr, ok := apperror.As(err); ok {
		if appErr.Status >= 500 {
			logger.ErrorContext(r.Context(), "request failed", "error", err)
		}
		write(w, appErr.Status, envelope{
			Success: false,
			Error:   errorBody{Code: appErr.Code, Message: appErr.Message, Fields: appErr.Fields},
		})
		return
	}
	logger.ErrorContext(r.Context(), "unhandled error", "error", err)
	write(w, http.StatusInternalServerError, envelope{
		Success: false,
		Error:   errorBody{Code: apperror.CodeInternal, Message: "Internal server error"},
	})
}
