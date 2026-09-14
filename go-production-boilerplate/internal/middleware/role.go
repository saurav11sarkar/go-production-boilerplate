package middleware

import (
	"log/slog"
	"net/http"

	"github.com/saurav11sarkar/go-production-boilerplate/internal/apperror"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/auth"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/domain"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/httpx"
)

func RequireRole(logger *slog.Logger, roles ...domain.Role) Middleware {
	allowed := make(map[domain.Role]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := auth.UserFromContext(r.Context())
			if !ok {
				httpx.Error(w, r, logger, apperror.Unauthorized("Authentication required"))
				return
			}
			if _, ok := allowed[user.Role]; !ok {
				httpx.Error(w, r, logger, apperror.Forbidden("You do not have permission to perform this action"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
