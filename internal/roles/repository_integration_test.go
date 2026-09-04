package roles

import (
	"context"
	"errors"
	"testing"

	"gatepass/internal/testutil"
)

func TestMembershipAndPermissionsRepository(t *testing.T) {
	db := testutil.OpenMySQL(t)
	for _, table := range []string{"tenant_memberships", "roles", "permissions", "role_permissions"} {
		testutil.RequireTable(t, db, table)
	}
	ctx := context.Background()

	var userID, roleID int64
	if err := db.QueryRowContext(ctx, `SELECT id, role_id FROM tenant_memberships WHERE status = 'active' LIMIT 1`).Scan(&userID, &roleID); err != nil {
		t.Skipf("no active membership available: %v", err)
	}

	r := NewRepository(db)
	m, err := r.MembershipFor(ctx, userID)
	if err != nil { t.Fatalf("MembershipFor: %v", err) }
	if m.UserID != userID || m.RoleID != roleID { t.Fatalf("unexpected membership: %+v", m) }
	if !m.IsActive() { t.Fatalf("membership should be active: %+v", m) }

	role, err := r.RoleByID(ctx, roleID)
	if err != nil { t.Fatalf("RoleByID: %v", err) }
	if role.ID != roleID { t.Fatalf("role ID = %d, want %d", role.ID, roleID) }

	codes, err := r.PermissionCodesForRole(ctx, roleID)
	if err != nil { t.Fatalf("PermissionCodesForRole: %v", err) }
	if codes == nil { t.Fatal("permission map must not be nil") }

	access, err := r.UserAccessByUserID(ctx, userID)
	if err != nil { t.Fatalf("UserAccessByUserID: %v", err) }
	if access.UserID != userID || access.RoleID != roleID { t.Fatalf("unexpected access: %+v", access) }
}

func TestMembershipMissingUserReturnsNotFound(t *testing.T) {
	db := testutil.OpenMySQL(t)
	testutil.RequireTable(t, db, "tenant_memberships")
	_, err := NewRepository(db).MembershipFor(context.Background(), -999999)
	if !errors.Is(err, ErrNotFound) { t.Fatalf("error = %v, want ErrNotFound", err) }
}
