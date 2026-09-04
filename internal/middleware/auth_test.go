package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gatepass/internal/auth"
	"gatepass/internal/tenants"
)

func TestRequireAuthAcceptsMatchingTenant(t *testing.T) {
	secret := []byte("middleware-secret")
	token, err := auth.IssueAccessToken(secret, 10, 20, 30, time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req = req.WithContext(WithTenant(req.Context(), &tenants.Tenant{ID: 20, Status: tenants.StatusActive}))

	called := false
	h := RequireAuth(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		claims, ok := ClaimsFromContext(r.Context())
		if !ok {
			t.Fatal("claims missing from context")
		}
		if claims.UserID != 10 || claims.TenantID != 20 || claims.RoleID != 30 {
			t.Fatalf("unexpected claims: %+v", claims)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if !called {
		t.Fatal("next handler was not called")
	}
	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNoContent)
	}
}

func TestRequireAuthRejectsMissingInvalidAndCrossTenantTokens(t *testing.T) {
	secret := []byte("middleware-secret")
	token, err := auth.IssueAccessToken(secret, 10, 20, 30, time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name      string
		header    string
		tenantID  int64
	}{
		{name: "missing bearer", header: "", tenantID: 20},
		{name: "invalid token", header: "Bearer invalid.token.value", tenantID: 20},
		{name: "cross tenant", header: "Bearer " + token, tenantID: 99},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			req = req.WithContext(WithTenant(req.Context(), &tenants.Tenant{ID: tc.tenantID, Status: tenants.StatusActive}))

			called := false
			h := RequireAuth(secret)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				called = true
			}))
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)

			if called {
				t.Fatal("next handler must not be called")
			}
			if rr.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
			}
		})
	}
}
