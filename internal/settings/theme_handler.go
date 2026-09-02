package settings

import (
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"gatepass/internal/httpx"
	"gatepass/internal/reqctx"
)

var hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

type ThemeHandler struct {
	repo *Repository
}

func NewThemeHandler(repo *Repository) *ThemeHandler {
	return &ThemeHandler{repo: repo}
}

type ThemeDTO struct {
	BrandName          string `json:"brand_name"`
	LogoURL            string `json:"logo_url"`
	FaviconURL         string `json:"favicon_url"`
	PrimaryColor       string `json:"primary_color"`
	PrimaryColorDark   string `json:"primary_color_dark"`
	AccentColor        string `json:"accent_color"`
	SidebarBackground  string `json:"sidebar_background"`
	SidebarText        string `json:"sidebar_text"`
	SidebarActiveBG    string `json:"sidebar_active_background"`
	SidebarActiveText  string `json:"sidebar_active_text"`
	Appearance         string `json:"appearance"`
}

func defaultTheme() ThemeDTO {
	return ThemeDTO{
		PrimaryColor: "#2563eb",
		PrimaryColorDark: "#1d4ed8",
		AccentColor: "#2563eb",
		SidebarBackground: "#ffffff",
		SidebarText: "#475467",
		SidebarActiveBG: "#eff6ff",
		SidebarActiveText: "#2563eb",
		Appearance: "light",
	}
}

func (h *ThemeHandler) current(r *http.Request) ThemeDTO {
	out := defaultTheme()
	raw, err := h.repo.Get(r.Context(), KeyTheme)
	if err != nil || len(raw) == 0 {
		return out
	}
	var saved ThemeDTO
	if json.Unmarshal(raw, &saved) != nil {
		return out
	}
	mergeTheme(&out, saved)
	return out
}

func mergeTheme(dst *ThemeDTO, src ThemeDTO) {
	if strings.TrimSpace(src.BrandName) != "" { dst.BrandName = strings.TrimSpace(src.BrandName) }
	if strings.TrimSpace(src.LogoURL) != "" { dst.LogoURL = strings.TrimSpace(src.LogoURL) }
	if strings.TrimSpace(src.FaviconURL) != "" { dst.FaviconURL = strings.TrimSpace(src.FaviconURL) }
	if hexColor.MatchString(src.PrimaryColor) { dst.PrimaryColor = src.PrimaryColor }
	if hexColor.MatchString(src.PrimaryColorDark) { dst.PrimaryColorDark = src.PrimaryColorDark }
	if hexColor.MatchString(src.AccentColor) { dst.AccentColor = src.AccentColor }
	if hexColor.MatchString(src.SidebarBackground) { dst.SidebarBackground = src.SidebarBackground }
	if hexColor.MatchString(src.SidebarText) { dst.SidebarText = src.SidebarText }
	if hexColor.MatchString(src.SidebarActiveBG) { dst.SidebarActiveBG = src.SidebarActiveBG }
	if hexColor.MatchString(src.SidebarActiveText) { dst.SidebarActiveText = src.SidebarActiveText }
	switch src.Appearance {
	case "light", "dark", "system":
		dst.Appearance = src.Appearance
	}
}

func (h *ThemeHandler) Get(w http.ResponseWriter, r *http.Request) {
	if _, ok := reqctx.TenantFromContext(r.Context()); !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, h.current(r))
}

func (h *ThemeHandler) Update(w http.ResponseWriter, r *http.Request) {
	if _, ok := reqctx.TenantFromContext(r.Context()); !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}
	claims, ok := reqctx.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrAuthRequired)
		return
	}

	var in ThemeDTO
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	normalized, validationMessage := validateTheme(in)
	if validationMessage != "" {
		httpx.WriteError(w, httpx.ErrValidation.WithMessage(validationMessage))
		return
	}
	if err := h.repo.Set(r.Context(), KeyTheme, normalized, claims.UserID); err != nil {
		httpx.WriteError(w, httpx.ErrInternal)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, normalized)
}

func validateTheme(in ThemeDTO) (ThemeDTO, string) {
	in.BrandName = strings.TrimSpace(in.BrandName)
	in.LogoURL = strings.TrimSpace(in.LogoURL)
	in.FaviconURL = strings.TrimSpace(in.FaviconURL)
	if len(in.BrandName) > 160 {
		return ThemeDTO{}, "brand_name must be 160 characters or fewer"
	}
	if message := validateAssetURL(in.LogoURL); message != "" {
		return ThemeDTO{}, "logo_url " + message
	}
	if message := validateAssetURL(in.FaviconURL); message != "" {
		return ThemeDTO{}, "favicon_url " + message
	}

	colors := []struct {
		name string
		value string
	}{
		{"primary_color", in.PrimaryColor},
		{"primary_color_dark", in.PrimaryColorDark},
		{"accent_color", in.AccentColor},
		{"sidebar_background", in.SidebarBackground},
		{"sidebar_text", in.SidebarText},
		{"sidebar_active_background", in.SidebarActiveBG},
		{"sidebar_active_text", in.SidebarActiveText},
	}
	for _, c := range colors {
		if !hexColor.MatchString(c.value) {
			return ThemeDTO{}, c.name + " must be a 6-digit hex color"
		}
	}
	switch in.Appearance {
	case "light", "dark", "system":
	default:
		return ThemeDTO{}, "appearance must be light, dark, or system"
	}
	return in, ""
}

func validateAssetURL(raw string) string {
	if raw == "" || strings.HasPrefix(raw, "/") {
		return ""
	}
	u, err := url.ParseRequestURI(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "must be an absolute http(s) URL or root-relative path"
	}
	return ""
}
