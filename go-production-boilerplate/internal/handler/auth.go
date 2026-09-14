package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/saurav11sarkar/go-production-boilerplate/internal/auth"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/config"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/dto"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/httpx"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/service"
)

type AuthHandler struct {
	service *service.AuthService
	logger  *slog.Logger
	cfg     config.Config
}

func NewAuthHandler(service *service.AuthService, logger *slog.Logger, cfg config.Config) *AuthHandler {
	return &AuthHandler{service: service, logger: logger, cfg: cfg}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.Error(w, r, h.logger, badJSON(err))
		return
	}
	user, err := h.service.Register(r.Context(), req)
	if err != nil {
		httpx.Error(w, r, h.logger, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, "Registration successful. Check your email for the verification link.", user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.Error(w, r, h.logger, badJSON(err))
		return
	}
	session, err := h.service.Login(r.Context(), req)
	if err != nil {
		httpx.Error(w, r, h.logger, err)
		return
	}
	h.setRefreshCookie(w, session.RefreshToken, session.RefreshExpiry)
	response := dto.AuthResponse{
		AccessToken: session.AccessToken,
		TokenType:   "Bearer",
		ExpiresAt:   session.AccessExpiry.Format(time.RFC3339),
	}
	if h.cfg.Auth.RefreshTokenInBody {
		response.RefreshToken = session.RefreshToken
	}
	httpx.JSON(w, http.StatusOK, "Login successful", response)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	raw := h.refreshTokenFromCookie(r)
	if raw == "" && h.cfg.Auth.RefreshTokenInBody {
		var req dto.RefreshRequest
		if err := httpx.DecodeJSON(w, r, &req); err == nil {
			raw = req.RefreshToken
		}
	}
	session, err := h.service.Refresh(r.Context(), raw)
	if err != nil {
		h.clearRefreshCookie(w)
		httpx.Error(w, r, h.logger, err)
		return
	}
	h.setRefreshCookie(w, session.RefreshToken, session.RefreshExpiry)
	response := dto.AuthResponse{
		AccessToken: session.AccessToken,
		TokenType:   "Bearer",
		ExpiresAt:   session.AccessExpiry.Format(time.RFC3339),
	}
	if h.cfg.Auth.RefreshTokenInBody {
		response.RefreshToken = session.RefreshToken
	}
	httpx.JSON(w, http.StatusOK, "Session refreshed", response)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	raw := h.refreshTokenFromCookie(r)
	if raw == "" && h.cfg.Auth.RefreshTokenInBody {
		var req dto.RefreshRequest
		if err := httpx.DecodeJSON(w, r, &req); err == nil {
			raw = req.RefreshToken
		}
	}
	if err := h.service.Logout(r.Context(), raw); err != nil {
		httpx.Error(w, r, h.logger, err)
		return
	}
	h.clearRefreshCookie(w)
	httpx.JSON(w, http.StatusOK, "Logged out successfully", nil)
}

func (h *AuthHandler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())
	if err := h.service.LogoutAll(r.Context(), user.UserID); err != nil {
		httpx.Error(w, r, h.logger, err)
		return
	}
	h.clearRefreshCookie(w)
	httpx.JSON(w, http.StatusOK, "Logged out from all sessions", nil)
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req dto.VerifyEmailRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.Error(w, r, h.logger, badJSON(err))
		return
	}
	if fields := req.Validate(); len(fields) > 0 {
		httpx.Error(w, r, h.logger, validation(fields))
		return
	}
	if err := h.service.VerifyEmail(r.Context(), req.Token); err != nil {
		httpx.Error(w, r, h.logger, err)
		return
	}
	httpx.JSON(w, http.StatusOK, "Email verified successfully", nil)
}

func (h *AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	var req dto.ForgotPasswordRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.Error(w, r, h.logger, badJSON(err))
		return
	}
	if fields := req.Validate(); len(fields) > 0 {
		httpx.Error(w, r, h.logger, validation(fields))
		return
	}
	if err := h.service.ResendVerification(r.Context(), req.Email); err != nil {
		httpx.Error(w, r, h.logger, err)
		return
	}
	httpx.JSON(w, http.StatusOK, "If the account exists and is not verified, a verification email has been sent", nil)
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req dto.ForgotPasswordRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.Error(w, r, h.logger, badJSON(err))
		return
	}
	if err := h.service.ForgotPassword(r.Context(), req); err != nil {
		httpx.Error(w, r, h.logger, err)
		return
	}
	httpx.JSON(w, http.StatusOK, "If the email exists, a password reset link has been sent", nil)
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req dto.ResetPasswordRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.Error(w, r, h.logger, badJSON(err))
		return
	}
	if err := h.service.ResetPassword(r.Context(), req); err != nil {
		httpx.Error(w, r, h.logger, err)
		return
	}
	h.clearRefreshCookie(w)
	httpx.JSON(w, http.StatusOK, "Password reset successful. Please log in again.", nil)
}

func (h *AuthHandler) setRefreshCookie(w http.ResponseWriter, token string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.cfg.Auth.RefreshCookieName,
		Value:    token,
		Path:     "/api/v1/auth",
		HttpOnly: true,
		Secure:   h.cfg.Auth.RefreshCookieSecure,
		SameSite: h.refreshCookieSameSite(),
		Domain:   h.cfg.Auth.RefreshCookieDomain,
		Expires:  expires,
		MaxAge:   int(time.Until(expires).Seconds()),
	})
}

func (h *AuthHandler) clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name: h.cfg.Auth.RefreshCookieName, Value: "", Path: "/api/v1/auth",
		HttpOnly: true, Secure: h.cfg.Auth.RefreshCookieSecure, SameSite: h.refreshCookieSameSite(),
		Domain: h.cfg.Auth.RefreshCookieDomain, Expires: time.Unix(0, 0), MaxAge: -1,
	})
}

func (h *AuthHandler) refreshTokenFromCookie(r *http.Request) string {
	cookie, err := r.Cookie(h.cfg.Auth.RefreshCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (h *AuthHandler) refreshCookieSameSite() http.SameSite {
	switch h.cfg.Auth.RefreshCookieSameSite {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}
