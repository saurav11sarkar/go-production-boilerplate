package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/database"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/domain"
)

type RefreshTokenRepository struct {
	db *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, token domain.RefreshToken) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, token_version, expires_at)
		VALUES ($1, $2, $3, $4)`, token.UserID, token.TokenHash, token.TokenVersion, token.ExpiresAt)
	return err
}

func (r *RefreshTokenRepository) Rotate(ctx context.Context, oldHash string, next domain.RefreshToken) (uuid.UUID, int, error) {
	var userID uuid.UUID
	var tokenVersion int
	err := database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		var expiresAt time.Time
		var revokedAt *time.Time
		var currentUserVersion int
		err := tx.QueryRowContext(ctx, `
			SELECT rt.user_id, rt.expires_at, rt.revoked_at, rt.token_version, u.token_version
			FROM refresh_tokens rt
			JOIN users u ON u.id = rt.user_id
			WHERE rt.token_hash = $1
			FOR UPDATE OF rt, u`, oldHash).Scan(&userID, &expiresAt, &revokedAt, &tokenVersion, &currentUserVersion)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidToken
		}
		if err != nil {
			return err
		}
		if revokedAt != nil || !expiresAt.After(time.Now().UTC()) || tokenVersion != currentUserVersion {
			return ErrInvalidToken
		}
		if _, err := tx.ExecContext(ctx,
			`UPDATE refresh_tokens SET revoked_at = NOW() WHERE token_hash = $1`, oldHash); err != nil {
			return err
		}
		next.UserID = userID
		next.TokenVersion = currentUserVersion
		_, err = tx.ExecContext(ctx, `
			INSERT INTO refresh_tokens (user_id, token_hash, token_version, expires_at)
			VALUES ($1, $2, $3, $4)`, next.UserID, next.TokenHash, next.TokenVersion, next.ExpiresAt)
		return err
	})
	return userID, tokenVersion, err
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, tokenHash string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = COALESCE(revoked_at, NOW())
		WHERE token_hash = $1`, tokenHash)
	return err
}

func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE refresh_tokens SET revoked_at = COALESCE(revoked_at, NOW())
		WHERE user_id = $1 AND revoked_at IS NULL`, userID)
	return err
}

func (r *RefreshTokenRepository) CleanupExpired(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE expires_at < NOW() - INTERVAL '7 days'`)
	return err
}
