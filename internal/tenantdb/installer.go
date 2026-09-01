package tenantdb

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gatepass/internal/migrations"

	_ "github.com/go-sql-driver/mysql"
)

type Installer struct {
	Dir string
}

func NewInstaller(dir string) *Installer {
	if dir == "" { dir = "migrations/tenant" }
	return &Installer{Dir: dir}
}

// Install runs the tenant migration set against one isolated tenant database.
// The tenant row is seeded locally after the schema exists using the platform
// tenant ID, preserving the current tenant_id repository contract during the
// database-per-tenant transition.
func (i *Installer) Install(ctx context.Context, creds Credentials, tenantID int64, tenantName, slug, domainToken string) error {
	if tenantID < 1 { return fmt.Errorf("tenantdb: invalid tenant id") }
	db, err := sql.Open("mysql", mysqlDSN(creds))
	if err != nil { return err }
	defer db.Close()
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil { return err }

	lockName := fmt.Sprintf("%s_tenant_%d", migrations.LockPrefix, tenantID)
	if err := migrations.RunUp(ctx, db, i.Dir, lockName); err != nil { return err }

	_, err = db.ExecContext(ctx, "INSERT INTO tenants (id, name, slug, status, custom_domain_token, created_at, updated_at) VALUES (?, ?, ?, 'active', ?, NOW(), NOW()) ON DUPLICATE KEY UPDATE name=VALUES(name), slug=VALUES(slug), status='active', custom_domain_token=VALUES(custom_domain_token), updated_at=NOW()", tenantID, tenantName, slug, domainToken)
	return err
}
