package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/domain"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/query"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	const q = `
		INSERT INTO users (name, email, password_hash, role, is_verified, is_active)
		VALUES ($1, LOWER($2), $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`
	err := r.db.QueryRowContext(ctx, q,
		user.Name, user.Email, user.PasswordHash, user.Role, user.IsVerified, user.IsActive,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if isUniqueViolation(err) {
		return ErrEmailExists
	}
	return err
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	const q = `
		SELECT id, name, email, password_hash, role, avatar_url, avatar_public_id,
		       is_verified, is_active, token_version, created_at, updated_at
		FROM users WHERE id = $1`
	return scanUser(r.db.QueryRowContext(ctx, q, id))
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	const q = `
		SELECT id, name, email, password_hash, role, avatar_url, avatar_public_id,
		       is_verified, is_active, token_version, created_at, updated_at
		FROM users WHERE LOWER(email) = LOWER($1)`
	return scanUser(r.db.QueryRowContext(ctx, q, email))
}

func (r *UserRepository) UpdateProfile(ctx context.Context, id uuid.UUID, name string) (*domain.User, error) {
	const q = `
		UPDATE users SET name = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, email, password_hash, role, avatar_url, avatar_public_id,
		          is_verified, is_active, token_version, created_at, updated_at`
	return scanUser(r.db.QueryRowContext(ctx, q, id, name))
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, hash string) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE users SET password_hash = $2, token_version = token_version + 1, updated_at = NOW() WHERE id = $1`, id, hash)
	if err != nil {
		return err
	}
	return requireAffected(result)
}

func (r *UserRepository) UpdateAvatar(ctx context.Context, id uuid.UUID, url, publicID *string) (*domain.User, error) {
	const q = `
		UPDATE users SET avatar_url = $2, avatar_public_id = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, email, password_hash, role, avatar_url, avatar_public_id,
		          is_verified, is_active, token_version, created_at, updated_at`
	return scanUser(r.db.QueryRowContext(ctx, q, id, url, publicID))
}

func (r *UserRepository) MarkVerified(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE users SET is_verified = TRUE, updated_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return err
	}
	return requireAffected(result)
}

func (r *UserRepository) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE users SET is_active = $2, updated_at = NOW() WHERE id = $1`, id, active)
	if err != nil {
		return err
	}
	return requireAffected(result)
}

func (r *UserRepository) SetRole(ctx context.Context, id uuid.UUID, role domain.Role) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE users SET role = $2, updated_at = NOW() WHERE id = $1`, id, role)
	if err != nil {
		return err
	}
	return requireAffected(result)
}

func (r *UserRepository) IncrementTokenVersion(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `UPDATE users SET token_version = token_version + 1, updated_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return err
	}
	return requireAffected(result)
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return err
	}
	return requireAffected(result)
}

func (r *UserRepository) List(ctx context.Context, p query.Params) ([]domain.User, int64, error) {
	where, args := buildUserFilters(p)

	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	sortColumns := map[string]string{
		"createdAt": "created_at",
		"updatedAt": "updated_at",
		"name":      "name",
		"email":     "email",
		"role":      "role",
	}
	sortColumn, ok := sortColumns[p.SortBy]
	if !ok {
		sortColumn = "created_at"
	}
	sortOrder := "DESC"
	if strings.EqualFold(p.SortOrder, "asc") {
		sortOrder = "ASC"
	}

	limitIndex := len(args) + 1
	offsetIndex := len(args) + 2
	q := fmt.Sprintf(`
		SELECT id, name, email, password_hash, role, avatar_url, avatar_public_id,
		       is_verified, is_active, token_version, created_at, updated_at
		FROM users %s
		ORDER BY %s %s, id ASC
		LIMIT $%d OFFSET $%d`, where, sortColumn, sortOrder, limitIndex, offsetIndex)
	args = append(args, p.Limit, p.Offset())

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users := make([]domain.User, 0, p.Limit)
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(
			&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role,
			&u.AvatarURL, &u.AvatarPublicID, &u.IsVerified, &u.IsActive, &u.TokenVersion,
			&u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func buildUserFilters(p query.Params) (string, []any) {
	conditions := []string{"1=1"}
	args := []any{}
	add := func(condition string, value any) {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf(condition, len(args)))
	}

	if p.Search != "" {
		args = append(args, "%"+p.Search+"%")
		n := len(args)
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR email ILIKE $%d)", n, n))
	}
	if p.Role != "" {
		add("role = $%d", p.Role)
	}
	if p.IsVerified != nil {
		add("is_verified = $%d", *p.IsVerified)
	}
	if p.IsActive != nil {
		add("is_active = $%d", *p.IsActive)
	}
	return "WHERE " + strings.Join(conditions, " AND "), args
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanUser(row rowScanner) (*domain.User, error) {
	var u domain.User
	err := row.Scan(
		&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role,
		&u.AvatarURL, &u.AvatarPublicID, &u.IsVerified, &u.IsActive, &u.TokenVersion,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func requireAffected(result sql.Result) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
