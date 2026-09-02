package media

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gatepass/internal/httpx"
	"gatepass/internal/reqctx"
)

const (
	defaultMaxUploadBytes int64 = 5 << 20
	maxNameLength               = 255
)

type Handler struct {
	repo          *Repository
	storageRoot   string
	publicBaseURL string
	maxUploadBytes int64
}

func NewHandler(repo *Repository, storageRoot, publicBaseURL string, maxUploadBytes int64) *Handler {
	if strings.TrimSpace(storageRoot) == "" { storageRoot = "storage/media" }
	if maxUploadBytes <= 0 { maxUploadBytes = defaultMaxUploadBytes }
	return &Handler{repo: repo, storageRoot: storageRoot, publicBaseURL: strings.TrimRight(strings.TrimSpace(publicBaseURL), "/"), maxUploadBytes: maxUploadBytes}
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	tenant, ok := reqctx.TenantFromContext(r.Context())
	if !ok { httpx.WriteError(w, httpx.ErrAuthRequired); return }
	claims, ok := reqctx.ClaimsFromContext(r.Context())
	if !ok { httpx.WriteError(w, httpx.ErrAuthRequired); return }

	r.Body = http.MaxBytesReader(w, r.Body, h.maxUploadBytes+1024*1024)
	if err := r.ParseMultipartForm(h.maxUploadBytes); err != nil {
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("upload is invalid or exceeds the maximum allowed size"))
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("file is required"))
		return
	}
	defer file.Close()

	head := make([]byte, 512)
	n, err := io.ReadFull(file, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("unable to read uploaded file"))
		return
	}
	head = head[:n]
	mimeType := http.DetectContentType(head)
	ext, allowed := allowedImage(mimeType)
	if !allowed {
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("only PNG, JPEG, GIF, WebP and ICO images are allowed"))
		return
	}

	publicID, err := randomID()
	if err != nil { httpx.WriteError(w, httpx.ErrInternal); return }
	relativePath := filepath.Join("tenant-"+itoa(tenant.ID), publicID+ext)
	absolutePath := filepath.Join(h.storageRoot, relativePath)
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0750); err != nil {
		httpx.WriteError(w, httpx.ErrInternal); return
	}

	dst, err := os.OpenFile(absolutePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0640)
	if err != nil { httpx.WriteError(w, httpx.ErrInternal); return }
	written, copyErr := io.Copy(dst, io.MultiReader(strings.NewReader(string(head)), file))
	closeErr := dst.Close()
	if copyErr != nil || closeErr != nil {
		_ = os.Remove(absolutePath)
		httpx.WriteError(w, httpx.ErrInternal)
		return
	}
	if written > h.maxUploadBytes {
		_ = os.Remove(absolutePath)
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("file exceeds the maximum allowed size"))
		return
	}

	name := filepath.Base(strings.TrimSpace(header.Filename))
	if name == "." || name == "" { name = "upload" + ext }
	if len(name) > maxNameLength { name = name[:maxNameLength] }
	purpose := strings.TrimSpace(r.FormValue("purpose"))
	if purpose == "" { purpose = "general" }
	if len(purpose) > 50 { purpose = purpose[:50] }

	f := &File{PublicID: publicID, OriginalName: name, StoragePath: filepath.ToSlash(relativePath), MimeType: mimeType, SizeBytes: written, Purpose: purpose, CreatedBy: claims.UserID}
	id, err := h.repo.Create(r.Context(), f)
	if err != nil {
		_ = os.Remove(absolutePath)
		httpx.WriteError(w, httpx.ErrInternal)
		return
	}
	f.ID = id
	httpx.WriteJSON(w, http.StatusCreated, h.response(r, f))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	files, err := h.repo.List(r.Context(), 100)
	if err != nil { httpx.WriteError(w, httpx.ErrInternal); return }
	out := make([]map[string]any, 0, len(files))
	for i := range files { out = append(out, h.response(r, &files[i])) }
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(strings.TrimSpace(r.PathValue("id")), 10, 64)
	if err != nil || id <= 0 {
		httpx.WriteError(w, httpx.ErrValidation.WithMessage("invalid media id"))
		return
	}
	f, err := h.repo.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) { httpx.WriteError(w, httpx.ErrNotFound); return }
		httpx.WriteError(w, httpx.ErrInternal); return
	}
	_ = os.Remove(filepath.Join(h.storageRoot, filepath.FromSlash(f.StoragePath)))
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Public(w http.ResponseWriter, r *http.Request) {
	publicID := strings.TrimSpace(r.PathValue("publicID"))
	if len(publicID) != 32 { httpx.WriteError(w, httpx.ErrNotFound); return }
	f, err := h.repo.ByPublicID(r.Context(), publicID)
	if err != nil {
		if errors.Is(err, ErrNotFound) { httpx.WriteError(w, httpx.ErrNotFound); return }
		httpx.WriteError(w, httpx.ErrInternal); return
	}
	path := filepath.Join(h.storageRoot, filepath.FromSlash(f.StoragePath))
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) { httpx.WriteError(w, httpx.ErrNotFound); return }
		httpx.WriteError(w, httpx.ErrInternal); return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil { httpx.WriteError(w, httpx.ErrInternal); return }
	w.Header().Set("Content-Type", f.MimeType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	http.ServeContent(w, r, f.OriginalName, info.ModTime(), file)
}

func (h *Handler) response(r *http.Request, f *File) map[string]any {
	return map[string]any{
		"id": f.ID, "public_id": f.PublicID, "original_name": f.OriginalName,
		"mime_type": f.MimeType, "size_bytes": f.SizeBytes, "purpose": f.Purpose,
		"created_by": f.CreatedBy, "created_at": f.CreatedAt,
		"public_url": h.publicURL(r, f.PublicID),
	}
}

func (h *Handler) publicURL(r *http.Request, publicID string) string {
	base := h.publicBaseURL
	if base == "" {
		scheme := "http"
		if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") { scheme = "https" }
		base = scheme + "://" + r.Host
	}
	return base + "/api/v1/media/public/" + publicID
}

func allowedImage(mimeType string) (string, bool) {
	switch mimeType {
	case "image/png": return ".png", true
	case "image/jpeg": return ".jpg", true
	case "image/gif": return ".gif", true
	case "image/webp": return ".webp", true
	case "image/x-icon", "image/vnd.microsoft.icon": return ".ico", true
	default: return "", false
	}
}

func randomID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil { return "", err }
	return hex.EncodeToString(b), nil
}

func itoa(v int64) string {
	const digits = "0123456789"
	if v == 0 { return "0" }
	var buf [20]byte
	i := len(buf)
	for v > 0 { i--; buf[i] = digits[v%10]; v /= 10 }
	return string(buf[i:])
}
