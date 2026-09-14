package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/saurav11sarkar/go-production-boilerplate/internal/auth"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/config"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/domain"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/handler"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/mailer"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/middleware"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/repository"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/security"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/service"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/storage"
)

type App struct {
	cfg    config.Config
	logger *slog.Logger
	db     *sql.DB
	server *http.Server
}

func New(cfg config.Config, logger *slog.Logger, db *sql.DB) (*App, error) {
	jwtManager := auth.NewJWTManager(cfg.Auth.Issuer, cfg.Auth.AccessTokenSecret, cfg.Auth.AccessTokenTTL)
	passwords := security.NewPasswordManager()
	storageProvider, err := storage.New(cfg)
	if err != nil {
		return nil, err
	}
	mailProvider, err := mailer.New(cfg, logger)
	if err != nil {
		return nil, err
	}

	userRepo := repository.NewUserRepository(db)
	refreshRepo := repository.NewRefreshTokenRepository(db)
	authTokenRepo := repository.NewAuthTokenRepository(db)

	authService := service.NewAuthService(
		userRepo, refreshRepo, authTokenRepo, jwtManager, passwords, mailProvider, logger,
		cfg.FrontendURL, cfg.Auth.RefreshTokenTTL, cfg.Auth.RequireEmailVerification,
	)
	userService := service.NewUserService(userRepo, refreshRepo, passwords, storageProvider, logger)
	adminService := service.NewAdminService(userRepo, refreshRepo, storageProvider, logger)

	healthHandler := handler.NewHealthHandler(db)
	authHandler := handler.NewAuthHandler(authService, logger, cfg)
	userHandler := handler.NewUserHandler(userService, logger, cfg.Storage.MaxImageBytes)
	adminHandler := handler.NewAdminHandler(adminService, logger)

	mux := http.NewServeMux()
	authRateLimiter := middleware.NewRateLimiter(cfg.HTTP.AuthRateLimitRequests, cfg.HTTP.AuthRateLimitWindow, cfg.HTTP.TrustedProxy)
	registerRoutes(mux, cfg, logger, jwtManager, userRepo, healthHandler, authHandler, userHandler, adminHandler, authRateLimiter.Middleware)

	rateLimiter := middleware.NewRateLimiter(cfg.HTTP.RateLimitRequests, cfg.HTTP.RateLimitWindow, cfg.HTTP.TrustedProxy)
	root := middleware.Chain(
		mux,
		middleware.RequestID,
		middleware.Recover(logger),
		middleware.Logger(logger, cfg.HTTP.TrustedProxy),
		middleware.SecurityHeaders(cfg.HTTP.TrustedProxy),
		middleware.CORS(cfg.HTTP.AllowedOrigins),
		rateLimiter.Middleware,
		middleware.Timeout(cfg.HTTP.RequestTimeout),
	)

	server := &http.Server{
		Addr:              ":" + cfg.HTTP.Port,
		Handler:           root,
		ReadHeaderTimeout: cfg.HTTP.ReadTimeout,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
	}

	return &App{cfg: cfg, logger: logger, db: db, server: server}, nil
}

func (a *App) Run() error {
	errCh := make(chan error, 1)
	go func() {
		a.logger.Info("server starting", "address", a.server.Addr, "environment", a.cfg.Environment)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-errCh:
		return fmt.Errorf("http server: %w", err)
	case <-sigCtx.Done():
		a.logger.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.cfg.HTTP.ShutdownTimeout)
	defer cancel()
	if err := a.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	a.logger.Info("server stopped cleanly")
	return nil
}

func adminOnly(jwtManager *auth.JWTManager, users middleware.AuthUserStore, logger *slog.Logger, h http.Handler) http.Handler {
	return middleware.Chain(h,
		middleware.RequireAuth(jwtManager, users, logger),
		middleware.RequireRole(logger, domain.RoleAdmin),
	)
}
