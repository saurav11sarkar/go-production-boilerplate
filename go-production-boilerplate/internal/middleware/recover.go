package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/saurav11sarkar/go-production-boilerplate/internal/apperror"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/httpx"
)

func Recover(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if value := recover(); value != nil {
					logger.ErrorContext(r.Context(), "panic recovered", "panic", value, "stack", string(debug.Stack()))
					httpx.Error(w, r, logger, apperror.Internal(nil))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
