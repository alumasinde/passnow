package gatedevices

import (
	"net/http"
	"strings"
	"time"

	"gatepass/internal/httpx"
	"gatepass/internal/reqctx"
)

type Handler struct {
	repo     *Repository
	verifier *Verifier
}

func NewHandler(r *Repository, v *Verifier) *Handler { return &Handler{repo: r, verifier: v} }

// ActivateResult embeds Device, so existing clients still see the same top-level fields.
type ActivateResult struct {
	Device
	DeviceToken string `json:"device_token"`
	ExpiresAt   string `json:"expires_at"`
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	if _, ok := reqctx.TenantFromContext(r.Context()); !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	v, e := h.repo.List(r.Context())
	if e != nil {
		httpx.WriteError(w, httpx.ErrInternal)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, v)
}

func (h *Handler) Activate(w http.ResponseWriter, r *http.Request) {
	tenant, ok := reqctx.TenantFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	var in ActivateInput
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	d, e := h.repo.Resolve(r.Context(), strings.TrimSpace(in.DeviceKey), strings.TrimSpace(in.DeviceSecret))
	if e != nil {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	tok, exp, e := h.verifier.IssueToken(d, tenant.ID)
	if e != nil {
		httpx.WriteError(w, httpx.ErrInternal)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, ActivateResult{Device: *d, DeviceToken: tok, ExpiresAt: exp.Format(time.RFC3339)})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	if _, ok := reqctx.TenantFromContext(r.Context()); !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	var in Input
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	in.DeviceKey = strings.ToLower(strings.TrimSpace(in.DeviceKey))
	if len(in.DeviceKey) < 24 || len(strings.TrimSpace(in.DeviceSecret)) < 24 || strings.TrimSpace(in.Name) == "" || in.GateID < 1 {
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("valid device_key, device_secret, name and gate_id are required"))
		return
	}
	d, e := h.repo.Create(r.Context(), in)
	if e != nil {
		httpx.WriteError(w, httpx.ErrValidation.WithMessage(e.Error()))
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, d)
}