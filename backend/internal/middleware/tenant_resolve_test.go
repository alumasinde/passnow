package middleware

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"gatepass/internal/tenants"
)

// A tiny fake SQL driver: "down" fails every statement, "empty" returns no rows.
type fakeDriver struct{ down bool }

func (d fakeDriver) Open(string) (driver.Conn, error) { return &fakeConn{down: d.down}, nil }

type fakeConn struct{ down bool }

func (c *fakeConn) Prepare(string) (driver.Stmt, error) {
	if c.down {
		return nil, errors.New("db down")
	}
	return fakeStmt{}, nil
}
func (c *fakeConn) Close() error              { return nil }
func (c *fakeConn) Begin() (driver.Tx, error) { return nil, errors.New("no tx") }

type fakeStmt struct{}

func (fakeStmt) Close() error  { return nil }
func (fakeStmt) NumInput() int { return -1 }
func (fakeStmt) Exec([]driver.Value) (driver.Result, error) {
	return nil, errors.New("no exec")
}
func (fakeStmt) Query([]driver.Value) (driver.Rows, error) { return emptyRows{}, nil }

type emptyRows struct{}

func (emptyRows) Columns() []string {
	return []string{"id", "name", "slug", "status", "custom_domain", "custom_domain_verified",
		"custom_domain_token", "created_at", "updated_at", "deleted_at"}
}
func (emptyRows) Close() error              { return nil }
func (emptyRows) Next([]driver.Value) error { return io.EOF }

func init() {
	sql.Register("faketenant_down", fakeDriver{down: true})
	sql.Register("faketenant_empty", fakeDriver{down: false})
}

func resolveStatus(t *testing.T, driverName, host string) int {
	t.Helper()
	db, err := sql.Open(driverName, "")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := ResolveTenant(tenants.NewRepository(db), "passnow.test")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
	req := httptest.NewRequest("GET", "/api/v1/me", nil)
	req.Host = host
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec.Code
}

func TestResolveTenantReports503WhenPlatformDBIsDown(t *testing.T) {
	for _, host := range []string{"acme.passnow.test", "gate.acme.com"} {
		if got := resolveStatus(t, "faketenant_down", host); got != http.StatusServiceUnavailable {
			t.Fatalf("host %s with DB down: got %d, want 503", host, got)
		}
	}
}

func TestResolveTenantReports404ForUnknownTenant(t *testing.T) {
	for _, host := range []string{"nobody.passnow.test", "unknown.example.org"} {
		if got := resolveStatus(t, "faketenant_empty", host); got != http.StatusNotFound {
			t.Fatalf("host %s unknown tenant: got %d, want 404", host, got)
		}
	}
}
