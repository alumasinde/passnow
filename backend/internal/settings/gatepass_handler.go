package settings

import (
	"net/http"

	"gatepass/internal/httpx"
	"gatepass/internal/reqctx"
)

type GatepassSettingsHandler struct {
	repo *Repository
}

func NewGatepassSettingsHandler(repo *Repository) *GatepassSettingsHandler {
	return &GatepassSettingsHandler{repo: repo}
}

type gatepassSettingsDTO struct {
	NumberPrefix  string `json:"number_prefix"`
	NumberUseYear bool   `json:"number_use_year"`
}

func (h *GatepassSettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	if _, ok := reqctx.TenantFromContext(r.Context()); !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, gatepassSettingsDTO{
		NumberPrefix:  h.repo.GetString(r.Context(), KeyGatepassNumberPrefix, "GP"),
		NumberUseYear: h.repo.GetBool(r.Context(), KeyGatepassNumberUseYear, true),
	})
}

func (h *GatepassSettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	if _, ok := reqctx.TenantFromContext(r.Context()); !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	claims, ok := reqctx.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	var in gatepassSettingsDTO
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	if in.NumberPrefix == "" {
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("number_prefix is required"))
		return
	}
	if err := h.repo.Set(r.Context(), KeyGatepassNumberPrefix, in.NumberPrefix, claims.UserID); err != nil {
		httpx.WriteError(w, httpx.ErrInternal)
		return
	}
	if err := h.repo.Set(r.Context(), KeyGatepassNumberUseYear, in.NumberUseYear, claims.UserID); err != nil {
		httpx.WriteError(w, httpx.ErrInternal)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, in)
}
