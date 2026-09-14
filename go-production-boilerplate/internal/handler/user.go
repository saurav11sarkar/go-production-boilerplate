package handler

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/saurav11sarkar/go-production-boilerplate/internal/apperror"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/auth"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/dto"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/httpx"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/service"
)

type UserHandler struct {
	service       *service.UserService
	logger        *slog.Logger
	maxImageBytes int64
}

func NewUserHandler(service *service.UserService, logger *slog.Logger, maxImageBytes int64) *UserHandler {
	return &UserHandler{service: service, logger: logger, maxImageBytes: maxImageBytes}
}

func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	current, _ := auth.UserFromContext(r.Context())
	user, err := h.service.GetMe(r.Context(), current.UserID)
	if err != nil {
		httpx.Error(w, r, h.logger, err)
		return
	}
	httpx.JSON(w, http.StatusOK, "Profile retrieved", user)
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateProfileRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.Error(w, r, h.logger, badJSON(err))
		return
	}
	current, _ := auth.UserFromContext(r.Context())
	user, err := h.service.UpdateProfile(r.Context(), current.UserID, req)
	if err != nil {
		httpx.Error(w, r, h.logger, err)
		return
	}
	httpx.JSON(w, http.StatusOK, "Profile updated", user)
}

func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req dto.ChangePasswordRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.Error(w, r, h.logger, badJSON(err))
		return
	}
	current, _ := auth.UserFromContext(r.Context())
	if err := h.service.ChangePassword(r.Context(), current.UserID, req); err != nil {
		httpx.Error(w, r, h.logger, err)
		return
	}
	httpx.JSON(w, http.StatusOK, "Password changed. Existing sessions were revoked; please log in again.", nil)
}

func (h *UserHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.maxImageBytes+(1<<20))
	if err := r.ParseMultipartForm(h.maxImageBytes); err != nil {
		httpx.Error(w, r, h.logger, apperror.BadRequest("Invalid multipart request or file too large"))
		return
	}
	file, header, err := r.FormFile("image")
	if err != nil {
		httpx.Error(w, r, h.logger, apperror.BadRequest("image file is required"))
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, h.maxImageBytes+1))
	if err != nil {
		httpx.Error(w, r, h.logger, apperror.Internal(err))
		return
	}
	if int64(len(data)) > h.maxImageBytes {
		httpx.Error(w, r, h.logger, apperror.BadRequest(fmt.Sprintf("Image must be <= %d MB", h.maxImageBytes/(1024*1024))))
		return
	}
	contentType := http.DetectContentType(data)
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
		httpx.Error(w, r, h.logger, apperror.BadRequest("Only JPEG, PNG, and WebP images are allowed"))
		return
	}
	current, _ := auth.UserFromContext(r.Context())
	user, err := h.service.UpdateAvatar(r.Context(), current.UserID, data, header.Filename, contentType)
	if err != nil {
		httpx.Error(w, r, h.logger, err)
		return
	}
	httpx.JSON(w, http.StatusOK, "Avatar updated", user)
}

func (h *UserHandler) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	current, _ := auth.UserFromContext(r.Context())
	user, err := h.service.DeleteAvatar(r.Context(), current.UserID)
	if err != nil {
		httpx.Error(w, r, h.logger, err)
		return
	}
	httpx.JSON(w, http.StatusOK, "Avatar removed", user)
}
