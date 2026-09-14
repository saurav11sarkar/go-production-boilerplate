package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/apperror"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/domain"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/query"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/repository"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/storage"
)

type AdminService struct {
	users         AdminUserStore
	refreshTokens RefreshTokenStore
	storage       storage.Storage
	logger        *slog.Logger
}

func NewAdminService(users AdminUserStore, refreshTokens RefreshTokenStore, storage storage.Storage, logger *slog.Logger) *AdminService {
	return &AdminService{users: users, refreshTokens: refreshTokens, storage: storage, logger: logger}
}

func (s *AdminService) ListUsers(ctx context.Context, p query.Params) ([]domain.User, query.Meta, error) {
	if p.Role != "" && !domain.Role(p.Role).Valid() {
		return nil, query.Meta{}, apperror.Validation(map[string]string{"role": "Role must be user or admin"})
	}
	users, total, err := s.users.List(ctx, p)
	if err != nil {
		return nil, query.Meta{}, apperror.Internal(err)
	}
	return users, query.NewMeta(p.Page, p.Limit, total), nil
}

func (s *AdminService) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user, err := s.users.GetByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperror.NotFound("User not found")
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return user, nil
}

func (s *AdminService) SetStatus(ctx context.Context, actorID, id uuid.UUID, active bool) error {
	if actorID == id && !active {
		return apperror.BadRequest("You cannot disable your own admin account")
	}
	if err := s.users.SetActive(ctx, id, active); errors.Is(err, repository.ErrNotFound) {
		return apperror.NotFound("User not found")
	} else if err != nil {
		return apperror.Internal(err)
	}
	if !active {
		if err := s.refreshTokens.RevokeAllForUser(ctx, id); err != nil {
			return apperror.Internal(err)
		}
	}
	return nil
}

func (s *AdminService) SetRole(ctx context.Context, actorID, id uuid.UUID, role domain.Role) error {
	if !role.Valid() {
		return apperror.Validation(map[string]string{"role": "Role must be user or admin"})
	}
	if actorID == id && role != domain.RoleAdmin {
		return apperror.BadRequest("You cannot remove your own admin role")
	}
	if err := s.users.SetRole(ctx, id, role); errors.Is(err, repository.ErrNotFound) {
		return apperror.NotFound("User not found")
	} else if err != nil {
		return apperror.Internal(err)
	}
	if err := s.refreshTokens.RevokeAllForUser(ctx, id); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *AdminService) DeleteUser(ctx context.Context, actorID, id uuid.UUID) error {
	if actorID == id {
		return apperror.BadRequest("You cannot delete your own admin account")
	}
	user, err := s.users.GetByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return apperror.NotFound("User not found")
	}
	if err != nil {
		return apperror.Internal(err)
	}
	if err := s.users.Delete(ctx, id); errors.Is(err, repository.ErrNotFound) {
		return apperror.NotFound("User not found")
	} else if err != nil {
		return apperror.Internal(err)
	}
	if user.AvatarPublicID != nil && *user.AvatarPublicID != "" {
		if err := s.storage.Delete(ctx, *user.AvatarPublicID); err != nil {
			s.logger.WarnContext(ctx, "failed to delete avatar for deleted user", "user_id", id, "public_id", *user.AvatarPublicID, "error", err)
		}
	}
	return nil
}
