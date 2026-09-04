package tenants

import (
	"testing"
	"time"
)

func TestTenantIsActive(t *testing.T) {
	active := &Tenant{Status: StatusActive}
	if !active.IsActive() {
		t.Fatal("active tenant should be active")
	}

	suspended := &Tenant{Status: StatusSuspended}
	if suspended.IsActive() {
		t.Fatal("suspended tenant should not be active")
	}

	now := time.Now()
	deleted := &Tenant{Status: StatusActive, DeletedAt: &now}
	if deleted.IsActive() {
		t.Fatal("soft-deleted tenant should not be active")
	}
}
