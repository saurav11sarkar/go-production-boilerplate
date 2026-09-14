package app

import (
	"log/slog"
	"net/http"

	"github.com/saurav11sarkar/go-production-boilerplate/internal/auth"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/config"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/handler"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/middleware"
)

func registerRoutes(
	mux *http.ServeMux,
	cfg config.Config,
	logger *slog.Logger,
	jwtManager *auth.JWTManager,
	users middleware.AuthUserStore,
	health *handler.HealthHandler,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	adminHandler *handler.AdminHandler,
	authSensitive middleware.Middleware,
) {
	mux.HandleFunc("GET /healthz", health.Liveness)
	mux.HandleFunc("GET /readyz", health.Readiness)

	mux.Handle("POST /api/v1/auth/register", authSensitive(http.HandlerFunc(authHandler.Register)))
	mux.Handle("POST /api/v1/auth/login", authSensitive(http.HandlerFunc(authHandler.Login)))
	mux.HandleFunc("POST /api/v1/auth/refresh", authHandler.Refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", authHandler.Logout)
	mux.HandleFunc("POST /api/v1/auth/verify-email", authHandler.VerifyEmail)
	mux.Handle("POST /api/v1/auth/resend-verification", authSensitive(http.HandlerFunc(authHandler.ResendVerification)))
	mux.Handle("POST /api/v1/auth/forgot-password", authSensitive(http.HandlerFunc(authHandler.ForgotPassword)))
	mux.Handle("POST /api/v1/auth/reset-password", authSensitive(http.HandlerFunc(authHandler.ResetPassword)))

	protected := func(h http.Handler) http.Handler {
		return middleware.Chain(h, middleware.RequireAuth(jwtManager, users, logger))
	}
	mux.Handle("POST /api/v1/auth/logout-all", protected(http.HandlerFunc(authHandler.LogoutAll)))
	mux.Handle("GET /api/v1/users/me", protected(http.HandlerFunc(userHandler.Me)))
	mux.Handle("PATCH /api/v1/users/me", protected(http.HandlerFunc(userHandler.UpdateProfile)))
	mux.Handle("PATCH /api/v1/users/me/password", protected(http.HandlerFunc(userHandler.ChangePassword)))
	mux.Handle("POST /api/v1/users/me/avatar", protected(http.HandlerFunc(userHandler.UploadAvatar)))
	mux.Handle("DELETE /api/v1/users/me/avatar", protected(http.HandlerFunc(userHandler.DeleteAvatar)))

	mux.Handle("GET /api/v1/admin/users", adminOnly(jwtManager, users, logger, http.HandlerFunc(adminHandler.ListUsers)))
	mux.Handle("GET /api/v1/admin/users/{id}", adminOnly(jwtManager, users, logger, http.HandlerFunc(adminHandler.GetUser)))
	mux.Handle("PATCH /api/v1/admin/users/{id}/status", adminOnly(jwtManager, users, logger, http.HandlerFunc(adminHandler.SetStatus)))
	mux.Handle("PATCH /api/v1/admin/users/{id}/role", adminOnly(jwtManager, users, logger, http.HandlerFunc(adminHandler.SetRole)))
	mux.Handle("DELETE /api/v1/admin/users/{id}", adminOnly(jwtManager, users, logger, http.HandlerFunc(adminHandler.DeleteUser)))

	if cfg.Storage.Provider == "local" {
		files := http.StripPrefix("/uploads/", http.FileServer(http.Dir(cfg.Storage.LocalDir)))
		mux.Handle("GET /uploads/", files)
	}
}
