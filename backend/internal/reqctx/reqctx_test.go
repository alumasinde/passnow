package reqctx

import (
	"context"
	"testing"

	"gatepass/internal/tenants"
)

func TestTenantContextRoundTrip(t *testing.T) {
	ctx := context.Background()
	if _, ok := TenantFromContext(ctx); ok {
		t.Fatal("unexpected tenant in empty context")
	}

	tenant := &tenants.Tenant{ID: 7, Slug: "acme", Status: tenants.StatusActive}
	ctx = WithTenant(ctx, tenant)
	got, ok := TenantFromContext(ctx)
	if !ok || got != tenant {
		t.Fatalf("tenant context round trip failed: got=%v ok=%v", got, ok)
	}
}

func TestClaimsContextRoundTrip(t *testing.T) {
	ctx := context.Background()
	if _, ok := ClaimsFromContext(ctx); ok {
		t.Fatal("unexpected claims in empty context")
	}

	want := Claims{UserID: 1, TenantID: 2, RoleID: 3}
	ctx = WithClaims(ctx, want)
	got, ok := ClaimsFromContext(ctx)
	if !ok || got != want {
		t.Fatalf("claims context round trip failed: got=%+v ok=%v", got, ok)
	}
}
