package service

import (
	"context"
	"errors"
	"fmt"
	"html"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/apperror"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/auth"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/domain"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/dto"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/mailer"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/repository"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/security"
)

type AuthService struct {
	users                    AuthUserStore
	refreshTokens            RefreshTokenStore
	authTokens               AuthTokenStore
	jwt                      *auth.JWTManager
	passwords                security.PasswordManager
	mailer                   mailer.Mailer
	logger                   *slog.Logger
	frontendURL              string
	refreshTTL               time.Duration
	requireEmailVerification bool
}

type Session struct {
	AccessToken   string
	AccessExpiry  time.Time
	RefreshToken  string
	RefreshExpiry time.Time
}

func NewAuthService(
	users AuthUserStore,
	refreshTokens RefreshTokenStore,
	authTokens AuthTokenStore,
	jwtManager *auth.JWTManager,
	passwords security.PasswordManager,
	mailer mailer.Mailer,
	logger *slog.Logger,
	frontendURL string,
	refreshTTL time.Duration,
	requireEmailVerification bool,
) *AuthService {
	return &AuthService{
		users:                    users,
		refreshTokens:            refreshTokens,
		authTokens:               authTokens,
		jwt:                      jwtManager,
		passwords:                passwords,
		mailer:                   mailer,
		logger:                   logger,
		frontendURL:              frontendURL,
		refreshTTL:               refreshTTL,
		requireEmailVerification: requireEmailVerification,
	}
}

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (*domain.User, error) {
	if fields := req.Validate(); len(fields) > 0 {
		return nil, apperror.Validation(fields)
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if _, err := s.users.GetByEmail(ctx, email); err == nil {
		return nil, apperror.Conflict("Email already exists")
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, apperror.Internal(err)
	}

	hash, err := s.passwords.Hash(req.Password)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	user := &domain.User{
		Name:         strings.TrimSpace(req.Name),
		Email:        email,
		PasswordHash: hash,
		Role:         domain.RoleUser,
		IsVerified:   false,
		IsActive:     true,
	}
	if err := s.users.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrEmailExists) {
			return nil, apperror.Conflict("Email already exists")
		}
		return nil, apperror.Internal(err)
	}

	if err := s.sendVerification(ctx, user); err != nil {
		s.logger.ErrorContext(ctx, "verification email failed after registration", "user_id", user.ID, "error", err)
	}
	return user, nil
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*Session, error) {
	if fields := req.Validate(); len(fields) > 0 {
		return nil, apperror.Validation(fields)
	}
	user, err := s.users.GetByEmail(ctx, strings.TrimSpace(req.Email))
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperror.Unauthorized("Invalid email or password")
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if err := s.passwords.Compare(user.PasswordHash, req.Password); err != nil {
		return nil, apperror.Unauthorized("Invalid email or password")
	}
	if !user.IsActive {
		return nil, apperror.Forbidden("Account is disabled")
	}
	if s.requireEmailVerification && !user.IsVerified {
		return nil, apperror.Forbidden("Verify your email before logging in")
	}
	return s.createSession(ctx, user)
}

func (s *AuthService) Refresh(ctx context.Context, rawRefresh string) (*Session, error) {
	if strings.TrimSpace(rawRefresh) == "" {
		return nil, apperror.Unauthorized("Refresh token is required")
	}
	nextRaw, err := security.RandomToken(32)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	nextExpiry := time.Now().UTC().Add(s.refreshTTL)
	userID, tokenVersion, err := s.refreshTokens.Rotate(ctx, security.HashToken(rawRefresh), domain.RefreshToken{
		TokenHash: security.HashToken(nextRaw),
		ExpiresAt: nextExpiry,
	})
	if errors.Is(err, repository.ErrInvalidToken) {
		return nil, apperror.Unauthorized("Invalid or expired refresh token")
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, apperror.Unauthorized("Invalid refresh token")
	}
	if !user.IsActive {
		_ = s.refreshTokens.RevokeAllForUser(ctx, user.ID)
		return nil, apperror.Forbidden("Account is disabled")
	}
	if user.TokenVersion != tokenVersion {
		_ = s.refreshTokens.Revoke(ctx, security.HashToken(nextRaw))
		return nil, apperror.Unauthorized("Session has been revoked")
	}
	accessToken, accessExpiry, err := s.jwt.CreateAccessToken(user.ID, user.Role, tokenVersion)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return &Session{
		AccessToken:   accessToken,
		AccessExpiry:  accessExpiry,
		RefreshToken:  nextRaw,
		RefreshExpiry: nextExpiry,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, rawRefresh string) error {
	if rawRefresh == "" {
		return nil
	}
	if err := s.refreshTokens.Revoke(ctx, security.HashToken(rawRefresh)); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *AuthService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	if err := s.users.IncrementTokenVersion(ctx, userID); err != nil {
		return apperror.Internal(err)
	}
	if err := s.refreshTokens.RevokeAllForUser(ctx, userID); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, rawToken string) error {
	if strings.TrimSpace(rawToken) == "" {
		return apperror.BadRequest("Verification token is required")
	}
	userID, err := s.authTokens.ConsumeEmailVerification(ctx, security.HashToken(rawToken))
	if errors.Is(err, repository.ErrInvalidToken) {
		return apperror.BadRequest("Verification token is invalid or expired")
	}
	if err != nil {
		return apperror.Internal(err)
	}
	if err := s.users.MarkVerified(ctx, userID); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *AuthService) ResendVerification(ctx context.Context, email string) error {
	user, err := s.users.GetByEmail(ctx, strings.TrimSpace(email))
	if errors.Is(err, repository.ErrNotFound) {
		return nil
	}
	if err != nil {
		return apperror.Internal(err)
	}
	if user.IsVerified {
		return nil
	}
	if err := s.sendVerification(ctx, user); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *AuthService) ForgotPassword(ctx context.Context, req dto.ForgotPasswordRequest) error {
	if fields := req.Validate(); len(fields) > 0 {
		return apperror.Validation(fields)
	}
	user, err := s.users.GetByEmail(ctx, strings.TrimSpace(req.Email))
	if errors.Is(err, repository.ErrNotFound) {
		return nil
	}
	if err != nil {
		return apperror.Internal(err)
	}

	raw, err := security.RandomToken(32)
	if err != nil {
		return apperror.Internal(err)
	}
	if err := s.authTokens.CreatePasswordReset(ctx, user.ID, security.HashToken(raw), time.Now().UTC().Add(20*time.Minute)); err != nil {
		return apperror.Internal(err)
	}
	link := fmt.Sprintf("%s/reset-password?token=%s", s.frontendURL, raw)
	message := mailer.Message{
		To:      user.Email,
		Subject: "Reset your password",
		Text:    "Password reset link: " + link,
		HTML:    fmt.Sprintf(`<p>Reset your password:</p><p><a href="%s">Reset password</a></p><p>This link expires in 20 minutes.</p>`, link),
	}
	if err := s.mailer.Send(ctx, message); err != nil {
		s.logger.ErrorContext(ctx, "password reset email failed", "user_id", user.ID, "error", err)
	}
	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, req dto.ResetPasswordRequest) error {
	if fields := req.Validate(); len(fields) > 0 {
		return apperror.Validation(fields)
	}
	userID, err := s.authTokens.ConsumePasswordReset(ctx, security.HashToken(req.Token))
	if errors.Is(err, repository.ErrInvalidToken) {
		return apperror.BadRequest("Password reset token is invalid or expired")
	}
	if err != nil {
		return apperror.Internal(err)
	}
	hash, err := s.passwords.Hash(req.NewPassword)
	if err != nil {
		return apperror.Internal(err)
	}
	if err := s.users.UpdatePassword(ctx, userID, hash); err != nil {
		return apperror.Internal(err)
	}
	if err := s.refreshTokens.RevokeAllForUser(ctx, userID); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *AuthService) createSession(ctx context.Context, user *domain.User) (*Session, error) {
	access, accessExpiry, err := s.jwt.CreateAccessToken(user.ID, user.Role, user.TokenVersion)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	rawRefresh, err := security.RandomToken(32)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	refreshExpiry := time.Now().UTC().Add(s.refreshTTL)
	if err := s.refreshTokens.Create(ctx, domain.RefreshToken{
		UserID:       user.ID,
		TokenHash:    security.HashToken(rawRefresh),
		TokenVersion: user.TokenVersion,
		ExpiresAt:    refreshExpiry,
	}); err != nil {
		return nil, apperror.Internal(err)
	}
	return &Session{
		AccessToken:   access,
		AccessExpiry:  accessExpiry,
		RefreshToken:  rawRefresh,
		RefreshExpiry: refreshExpiry,
	}, nil
}

func (s *AuthService) sendVerification(ctx context.Context, user *domain.User) error {
	raw, err := security.RandomToken(32)
	if err != nil {
		return err
	}
	if err := s.authTokens.CreateEmailVerification(ctx, user.ID, security.HashToken(raw), time.Now().UTC().Add(24*time.Hour)); err != nil {
		return err
	}
	link := fmt.Sprintf("%s/verify-email?token=%s", s.frontendURL, raw)
	return s.mailer.Send(ctx, mailer.Message{
		To:      user.Email,
		Subject: "Verify your email",
		Text:    "Email verification link: " + link,
		HTML:    fmt.Sprintf(`<p>Welcome %s!</p><p><a href="%s">Verify your email</a></p><p>This link expires in 24 hours.</p>`, html.EscapeString(user.Name), link),
	})
}
