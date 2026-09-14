package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/domain"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/query"
)

type AuthUserStore interface {
	Create(context.Context, *domain.User) error
	GetByID(context.Context, uuid.UUID) (*domain.User, error)
	GetByEmail(context.Context, string) (*domain.User, error)
	MarkVerified(context.Context, uuid.UUID) error
	UpdatePassword(context.Context, uuid.UUID, string) error
	IncrementTokenVersion(context.Context, uuid.UUID) error
}

type RefreshTokenStore interface {
	Create(context.Context, domain.RefreshToken) error
	Rotate(context.Context, string, domain.RefreshToken) (uuid.UUID, int, error)
	Revoke(context.Context, string) error
	RevokeAllForUser(context.Context, uuid.UUID) error
}

type AuthTokenStore interface {
	CreateEmailVerification(context.Context, uuid.UUID, string, time.Time) error
	ConsumeEmailVerification(context.Context, string) (uuid.UUID, error)
	CreatePasswordReset(context.Context, uuid.UUID, string, time.Time) error
	ConsumePasswordReset(context.Context, string) (uuid.UUID, error)
}

type UserStore interface {
	GetByID(context.Context, uuid.UUID) (*domain.User, error)
	UpdateProfile(context.Context, uuid.UUID, string) (*domain.User, error)
	UpdatePassword(context.Context, uuid.UUID, string) error
	UpdateAvatar(context.Context, uuid.UUID, *string, *string) (*domain.User, error)
}

type AdminUserStore interface {
	GetByID(context.Context, uuid.UUID) (*domain.User, error)
	List(context.Context, query.Params) ([]domain.User, int64, error)
	SetActive(context.Context, uuid.UUID, bool) error
	SetRole(context.Context, uuid.UUID, domain.Role) error
	Delete(context.Context, uuid.UUID) error
}
