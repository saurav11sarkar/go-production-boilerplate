package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/database"
)

type AuthTokenRepository struct {
	db *sql.DB
}

func NewAuthTokenRepository(db *sql.DB) *AuthTokenRepository {
	return &AuthTokenRepository{db: db}
}

func (r *AuthTokenRepository) CreateEmailVerification(ctx context.Context, userID uuid.UUID, hash string, expiresAt time.Time) error {
	return database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `DELETE FROM email_verification_tokens WHERE user_id = $1 AND used_at IS NULL`, userID); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO email_verification_tokens (user_id, token_hash, expires_at)
			VALUES ($1, $2, $3)`, userID, hash, expiresAt)
		return err
	})
}

func (r *AuthTokenRepository) ConsumeEmailVerification(ctx context.Context, hash string) (uuid.UUID, error) {
	return r.consume(ctx, "email_verification_tokens", hash)
}

func (r *AuthTokenRepository) CreatePasswordReset(ctx context.Context, userID uuid.UUID, hash string, expiresAt time.Time) error {
	return database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `DELETE FROM password_reset_tokens WHERE user_id = $1 AND used_at IS NULL`, userID); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
			VALUES ($1, $2, $3)`, userID, hash, expiresAt)
		return err
	})
}

func (r *AuthTokenRepository) ConsumePasswordReset(ctx context.Context, hash string) (uuid.UUID, error) {
	return r.consume(ctx, "password_reset_tokens", hash)
}

func (r *AuthTokenRepository) consume(ctx context.Context, table, hash string) (uuid.UUID, error) {
	var userID uuid.UUID
	err := database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		var expiresAt time.Time
		var usedAt *time.Time
		query := `SELECT user_id, expires_at, used_at FROM ` + table + ` WHERE token_hash = $1 FOR UPDATE`
		err := tx.QueryRowContext(ctx, query, hash).Scan(&userID, &expiresAt, &usedAt)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidToken
		}
		if err != nil {
			return err
		}
		if usedAt != nil || !expiresAt.After(time.Now().UTC()) {
			return ErrInvalidToken
		}
		_, err = tx.ExecContext(ctx, `UPDATE `+table+` SET used_at = NOW() WHERE token_hash = $1`, hash)
		return err
	})
	return userID, err
}
