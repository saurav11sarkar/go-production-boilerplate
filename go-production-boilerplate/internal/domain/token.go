package domain

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	TokenHash    string
	TokenVersion int
	ExpiresAt    time.Time
	RevokedAt    *time.Time
	CreatedAt    time.Time
}

type OneTimeToken struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	TokenHash    string
	TokenVersion int
	ExpiresAt    time.Time
	UsedAt       *time.Time
	CreatedAt    time.Time
}
