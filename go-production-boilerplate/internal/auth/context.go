package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/domain"
)

type contextKey string

const authUserKey contextKey = "auth-user"

type UserContext struct {
	UserID uuid.UUID
	Role   domain.Role
}

func WithUser(ctx context.Context, user UserContext) context.Context {
	return context.WithValue(ctx, authUserKey, user)
}

func UserFromContext(ctx context.Context) (UserContext, bool) {
	user, ok := ctx.Value(authUserKey).(UserContext)
	return user, ok
}
