package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gatepass/internal/auth"
	"gatepass/internal/reqctx"
	"gatepass/internal/roles"
	"gatepass/internal/testutil"
	"gatepass/internal/tenants"
	"gatepass/internal/users"
)

const sprint4JWTSecret = "sprint4-http-test-secret"

type sprint4TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

func newSprint4AuthMux(t *testing.T) (*http.ServeMux, *tenants.Tenant, int64) {
	t.Helper()

	db := testutil.OpenMySQL(t)
	for _, table := range []string{"users", "tenant_memberships", "roles", "refresh_tokens"} {
		testutil.RequireTable(t, db, table)
	}

	ctx := context.Background()
	var roleID int64
	if err := db.QueryRowContext(ctx, `SELECT role_id FROM tenant_memberships WHERE status='active' LIMIT 1`).Scan(&roleID); err != nil {
		t.Skipf("no active role membership fixture available: %v", err)
	}

	password := "Sprint4-Test-Password-2026"
	hash, err := auth.HashPassword(password, 4)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	uniq := time.Now().UnixNano()
	userRepo := users.NewRepository(db)
	userID, err := userRepo.Create(ctx, &users.User{
		Email:        fmt.Sprintf("sprint4-%d@example.test", uniq),
		PasswordHash: hash,
		FirstName:    "Sprint",
		LastName:     "Four",
		Status:       users.StatusActive,
	})
	if err != nil {
		t.Fatalf("create sprint 4 user: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE user_id=?`, userID)
		_, _ = db.ExecContext(ctx, `DELETE FROM tenant_memberships WHERE user_id=?`, userID)
		_, _ = db.ExecContext(ctx, `DELETE FROM users WHERE id=?`, userID)
	})

	if _, err := db.ExecContext(ctx, `INSERT INTO tenant_memberships(user_id,role_id,status) VALUES(?,?, 'active')`, userID, roleID); err != nil {
		t.Fatalf("create sprint 4 membership: %v", err)
	}

	roleRepo := roles.NewRepository(db)
	refreshRepo := auth.NewRefreshTokenRepository(db)
	authSvc := auth.NewService(userRepo, roleRepo, refreshRepo, []byte(sprint4JWTSecret), 4, 15*time.Minute, time.Hour)

	api := NewAPI([]byte(sprint4JWTSecret), roleRepo)
	api.AuthHandler = auth.NewHandler(authSvc)

	mux := http.NewServeMux()
	RegisterAPI(mux, api)

	tenant := &tenants.Tenant{
		ID:     1001,
		Name:   "Sprint 4 Tenant",
		Slug:   "sprint4",
		Status: tenants.StatusActive,
	}
	return mux, tenant, userID
}

func sprint4Request(method, path string, body any, tenant *tenants.Tenant) *http.Request {
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		raw, err := json.Marshal(body)
		if err != nil {
			panic(err)
		}
		reader = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if tenant != nil {
		req = req.WithContext(reqctx.WithTenant(req.Context(), tenant))
	}
	return req
}

func TestSprint4AuthHTTPLoginRefreshLogoutLifecycle(t *testing.T) {
	mux, tenant, userID := newSprint4AuthMux(t)

	var email string
	// The test fixture user is the only user created after this mux setup with
	// the Sprint 4 prefix; retrieve it from the authenticated account setup.
	// This keeps the HTTP request independent from repository internals.
	if err := testutil.OpenMySQL(t).QueryRowContext(context.Background(), `SELECT email FROM users WHERE id=?`, userID).Scan(&email); err != nil {
		t.Fatalf("load sprint 4 user email: %v", err)
	}

	loginBody := map[string]string{
		"email":    email,
		"password": "Sprint4-Test-Password-2026",
	}
	loginRec := httptest.NewRecorder()
	mux.ServeHTTP(loginRec, sprint4Request(http.MethodPost, "/api/v1/auth/login", loginBody, tenant))
	if loginRec.Code != http.StatusOK {
		t.Fatalf("login status=%d body=%s", loginRec.Code, loginRec.Body.String())
	}

	var login sprint4TokenResponse
	if err := json.NewDecoder(loginRec.Body).Decode(&login); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if login.AccessToken == "" || login.RefreshToken == "" || login.ExpiresIn <= 0 {
		t.Fatalf("invalid login token response: %+v", login)
	}

	meReq := sprint4Request(http.MethodGet, "/api/v1/auth/me", nil, tenant)
	meReq.Header.Set("Authorization", "Bearer "+login.AccessToken)
	meRec := httptest.NewRecorder()
	mux.ServeHTTP(meRec, meReq)
	if meRec.Code != http.StatusOK {
		t.Fatalf("me status=%d body=%s", meRec.Code, meRec.Body.String())
	}

	refreshRec := httptest.NewRecorder()
	mux.ServeHTTP(refreshRec, sprint4Request(http.MethodPost, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": login.RefreshToken,
	}, tenant))
	if refreshRec.Code != http.StatusOK {
		t.Fatalf("refresh status=%d body=%s", refreshRec.Code, refreshRec.Body.String())
	}

	var refreshed sprint4TokenResponse
	if err := json.NewDecoder(refreshRec.Body).Decode(&refreshed); err != nil {
		t.Fatalf("decode refresh response: %v", err)
	}
	if refreshed.AccessToken == "" || refreshed.RefreshToken == "" {
		t.Fatalf("invalid refresh response: %+v", refreshed)
	}

	logoutReq := sprint4Request(http.MethodPost, "/api/v1/auth/logout", nil, tenant)
	logoutReq.Header.Set("Authorization", "Bearer "+refreshed.AccessToken)
	logoutRec := httptest.NewRecorder()
	mux.ServeHTTP(logoutRec, logoutReq)
	if logoutRec.Code != http.StatusNoContent {
		t.Fatalf("logout status=%d body=%s", logoutRec.Code, logoutRec.Body.String())
	}

	reuseRec := httptest.NewRecorder()
	mux.ServeHTTP(reuseRec, sprint4Request(http.MethodPost, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": refreshed.RefreshToken,
	}, tenant))
	if reuseRec.Code != http.StatusUnauthorized {
		t.Fatalf("revoked refresh status=%d body=%s", reuseRec.Code, reuseRec.Body.String())
	}
}

func TestSprint4HTTPRejectsMissingTenantAndCrossTenantToken(t *testing.T) {
	mux, tenant, userID := newSprint4AuthMux(t)

	missingTenantRec := httptest.NewRecorder()
	mux.ServeHTTP(missingTenantRec, sprint4Request(http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email": "nobody@example.test", "password": "irrelevant",
	}, nil))
	if missingTenantRec.Code != http.StatusUnauthorized {
		t.Fatalf("missing tenant login status=%d body=%s", missingTenantRec.Code, missingTenantRec.Body.String())
	}

	token, err := auth.IssueAccessToken([]byte(sprint4JWTSecret), userID, tenant.ID+1, 1, time.Minute)
	if err != nil {
		t.Fatalf("issue cross-tenant token: %v", err)
	}
	req := sprint4Request(http.MethodGet, "/api/v1/auth/me", nil, tenant)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("cross-tenant token status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestSprint4ProtectedRouteRejectsUnauthenticatedRequests(t *testing.T) {
	mux, tenant, _ := newSprint4AuthMux(t)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, sprint4Request(http.MethodGet, "/api/v1/visitors", nil, tenant))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("protected route status=%d body=%s", rec.Code, rec.Body.String())
	}
}
