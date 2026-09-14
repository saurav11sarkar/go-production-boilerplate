package service

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/apperror"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/domain"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/dto"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/repository"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/security"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/storage"
)

type UserService struct {
	users         UserStore
	refreshTokens RefreshTokenStore
	passwords     security.PasswordManager
	storage       storage.Storage
	logger        *slog.Logger
}

func NewUserService(users UserStore, refreshTokens RefreshTokenStore, passwords security.PasswordManager, storage storage.Storage, logger *slog.Logger) *UserService {
	return &UserService{users: users, refreshTokens: refreshTokens, passwords: passwords, storage: storage, logger: logger}
}

func (s *UserService) GetMe(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := s.users.GetByID(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperror.NotFound("User not found")
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return user, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, req dto.UpdateProfileRequest) (*domain.User, error) {
	if fields := req.Validate(); len(fields) > 0 {
		return nil, apperror.Validation(fields)
	}
	user, err := s.users.UpdateProfile(ctx, userID, strings.TrimSpace(req.Name))
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperror.NotFound("User not found")
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return user, nil
}

func (s *UserService) ChangePassword(ctx context.Context, userID uuid.UUID, req dto.ChangePasswordRequest) error {
	if fields := req.Validate(); len(fields) > 0 {
		return apperror.Validation(fields)
	}
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return apperror.NotFound("User not found")
	}
	if err := s.passwords.Compare(user.PasswordHash, req.CurrentPassword); err != nil {
		return apperror.Unauthorized("Current password is incorrect")
	}
	newHash, err := s.passwords.Hash(req.NewPassword)
	if err != nil {
		return apperror.Internal(err)
	}
	if err := s.users.UpdatePassword(ctx, userID, newHash); err != nil {
		return apperror.Internal(err)
	}
	if err := s.refreshTokens.RevokeAllForUser(ctx, userID); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *UserService) UpdateAvatar(ctx context.Context, userID uuid.UUID, data []byte, filename, contentType string) (*domain.User, error) {
	current, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, apperror.NotFound("User not found")
	}
	upload, err := s.storage.UploadImage(ctx, bytes.NewReader(data), filename, contentType)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	url, publicID := upload.URL, upload.PublicID
	updated, err := s.users.UpdateAvatar(ctx, userID, &url, &publicID)
	if err != nil {
		_ = s.storage.Delete(ctx, upload.PublicID)
		return nil, apperror.Internal(err)
	}
	if current.AvatarPublicID != nil && *current.AvatarPublicID != "" {
		if err := s.storage.Delete(ctx, *current.AvatarPublicID); err != nil {
			s.logger.WarnContext(ctx, "failed to delete old avatar", "public_id", *current.AvatarPublicID, "error", err)
		}
	}
	return updated, nil
}

func (s *UserService) DeleteAvatar(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	current, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, apperror.NotFound("User not found")
	}
	updated, err := s.users.UpdateAvatar(ctx, userID, nil, nil)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if current.AvatarPublicID != nil && *current.AvatarPublicID != "" {
		if err := s.storage.Delete(ctx, *current.AvatarPublicID); err != nil {
			s.logger.WarnContext(ctx, "failed to delete avatar after database update", "public_id", *current.AvatarPublicID, "error", err)
		}
	}
	return updated, nil
}
