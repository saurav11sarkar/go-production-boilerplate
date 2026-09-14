package dto

import (
	"net/mail"
	"strings"
)

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r RegisterRequest) Validate() map[string]string {
	fields := map[string]string{}
	if len(strings.TrimSpace(r.Name)) < 2 {
		fields["name"] = "Name must be at least 2 characters"
	}
	email := strings.TrimSpace(r.Email)
	if addr, err := mail.ParseAddress(email); err != nil || !strings.EqualFold(addr.Address, email) {
		fields["email"] = "A valid email is required"
	}
	if len(r.Password) < 8 {
		fields["password"] = "Password must be at least 8 characters"
	}
	if len(r.Password) > 72 {
		fields["password"] = "Password must be at most 72 characters"
	}
	return fields
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r LoginRequest) Validate() map[string]string {
	fields := map[string]string{}
	email := strings.TrimSpace(r.Email)
	if addr, err := mail.ParseAddress(email); err != nil || !strings.EqualFold(addr.Address, email) {
		fields["email"] = "A valid email is required"
	}
	if r.Password == "" {
		fields["password"] = "Password is required"
	}
	return fields
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type VerifyEmailRequest struct {
	Token string `json:"token"`
}

func (r VerifyEmailRequest) Validate() map[string]string {
	fields := map[string]string{}
	if strings.TrimSpace(r.Token) == "" {
		fields["token"] = "Token is required"
	}
	return fields
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

func (r ForgotPasswordRequest) Validate() map[string]string {
	fields := map[string]string{}
	email := strings.TrimSpace(r.Email)
	if addr, err := mail.ParseAddress(email); err != nil || !strings.EqualFold(addr.Address, email) {
		fields["email"] = "A valid email is required"
	}
	return fields
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}

func (r ResetPasswordRequest) Validate() map[string]string {
	fields := map[string]string{}
	if strings.TrimSpace(r.Token) == "" {
		fields["token"] = "Token is required"
	}
	if len(r.NewPassword) < 8 {
		fields["newPassword"] = "Password must be at least 8 characters"
	}
	if len(r.NewPassword) > 72 {
		fields["newPassword"] = "Password must be at most 72 characters"
	}
	return fields
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

func (r ChangePasswordRequest) Validate() map[string]string {
	fields := map[string]string{}
	if r.CurrentPassword == "" {
		fields["currentPassword"] = "Current password is required"
	}
	if len(r.NewPassword) < 8 {
		fields["newPassword"] = "New password must be at least 8 characters"
	}
	if len(r.NewPassword) > 72 {
		fields["newPassword"] = "New password must be at most 72 characters"
	}
	return fields
}

type AuthResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken,omitempty"`
	TokenType    string `json:"tokenType"`
	ExpiresAt    string `json:"expiresAt"`
}
