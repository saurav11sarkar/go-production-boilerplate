package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/apperror"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/auth"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/domain"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/httpx"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/repository"
)

type AuthUserStore interface {
	GetByID(context.Context, uuid.UUID) (*domain.User, error)
}

func RequireAuth(jwtManager *auth.JWTManager, users AuthUserStore, logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := strings.TrimSpace(r.Header.Get("Authorization"))
			parts := strings.Fields(header)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				httpx.Error(w, r, logger, apperror.Unauthorized("Bearer access token is required"))
				return
			}
			userID, _, tokenVersion, err := jwtManager.ParseAccessToken(parts[1])
			if err != nil {
				httpx.Error(w, r, logger, apperror.Unauthorized("Access token is invalid or expired"))
				return
			}

			user, err := users.GetByID(r.Context(), userID)
			if errors.Is(err, repository.ErrNotFound) {
				httpx.Error(w, r, logger, apperror.Unauthorized("User no longer exists"))
				return
			}
			if err != nil {
				httpx.Error(w, r, logger, apperror.Internal(err))
				return
			}
			if !user.IsActive {
				httpx.Error(w, r, logger, apperror.Forbidden("Account is disabled"))
				return
			}
			if user.TokenVersion != tokenVersion {
				httpx.Error(w, r, logger, apperror.Unauthorized("Access token has been revoked"))
				return
			}

			ctx := auth.WithUser(r.Context(), auth.UserContext{UserID: user.ID, Role: user.Role})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
