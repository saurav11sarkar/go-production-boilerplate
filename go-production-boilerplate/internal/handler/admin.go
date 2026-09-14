package handler

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/apperror"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/auth"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/domain"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/dto"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/httpx"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/query"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/service"
)

type AdminHandler struct {
	service *service.AdminService
	logger  *slog.Logger
}

func NewAdminHandler(service *service.AdminService, logger *slog.Logger) *AdminHandler {
	return &AdminHandler{service: service, logger: logger}
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	params := query.Parse(r.URL.Query())
	users, meta, err := h.service.ListUsers(r.Context(), params)
	if err != nil {
		httpx.Error(w, r, h.logger, err)
		return
	}
	httpx.JSONWithMeta(w, http.StatusOK, "Users retrieved", users, meta)
}

func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpx.Error(w, r, h.logger, apperror.BadRequest("Invalid user ID"))
		return
	}
	user, err := h.service.GetUser(r.Context(), id)
	if err != nil {
		httpx.Error(w, r, h.logger, err)
		return
	}
	httpx.JSON(w, http.StatusOK, "User retrieved", user)
}

func (h *AdminHandler) SetStatus(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpx.Error(w, r, h.logger, apperror.BadRequest("Invalid user ID"))
		return
	}
	var req dto.AdminStatusRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.Error(w, r, h.logger, badJSON(err))
		return
	}
	actor, _ := auth.UserFromContext(r.Context())
	if err := h.service.SetStatus(r.Context(), actor.UserID, id, req.IsActive); err != nil {
		httpx.Error(w, r, h.logger, err)
		return
	}
	httpx.JSON(w, http.StatusOK, "User status updated", nil)
}

func (h *AdminHandler) SetRole(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpx.Error(w, r, h.logger, apperror.BadRequest("Invalid user ID"))
		return
	}
	var req dto.AdminRoleRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.Error(w, r, h.logger, badJSON(err))
		return
	}
	actor, _ := auth.UserFromContext(r.Context())
	if err := h.service.SetRole(r.Context(), actor.UserID, id, domain.Role(req.Role)); err != nil {
		httpx.Error(w, r, h.logger, err)
		return
	}
	httpx.JSON(w, http.StatusOK, "User role updated", nil)
}

func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpx.Error(w, r, h.logger, apperror.BadRequest("Invalid user ID"))
		return
	}
	actor, _ := auth.UserFromContext(r.Context())
	if err := h.service.DeleteUser(r.Context(), actor.UserID, id); err != nil {
		httpx.Error(w, r, h.logger, err)
		return
	}
	httpx.NoContent(w)
}
