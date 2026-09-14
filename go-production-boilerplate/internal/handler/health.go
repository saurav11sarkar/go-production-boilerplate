package handler

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/saurav11sarkar/go-production-boilerplate/internal/httpx"
)

type HealthHandler struct {
	db *sql.DB
}

func NewHealthHandler(db *sql.DB) *HealthHandler { return &HealthHandler{db: db} }

func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	httpx.JSON(w, http.StatusOK, "Service is alive", map[string]string{"status": "ok"})
}

func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := h.db.PingContext(ctx); err != nil {
		httpx.JSON(w, http.StatusServiceUnavailable, "Service is not ready", map[string]string{"database": "unavailable"})
		return
	}
	httpx.JSON(w, http.StatusOK, "Service is ready", map[string]string{"database": "ok"})
}
