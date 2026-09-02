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
	if dir == "" {
		dir = "migrations/tenant"
	}
	return &Installer{Dir: dir}
}

// Install runs the tenant migration set against one isolated tenant database.
//
// Tenant identity belongs exclusively to the platform database. A tenant
// database therefore receives only tenant-owned schema and seed data; it must
// never contain a tenants table or a tenant_id row-scoping contract.
func (i *Installer) Install(ctx context.Context, creds Credentials, tenantID int64, tenantName, slug, domainToken string) error {
	if tenantID < 1 {
		return fmt.Errorf("tenantdb: invalid tenant id")
	}

	db, err := sql.Open("mysql", mysqlDSN(creds))
	if err != nil {
		return err
	}
	defer db.Close()

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		return err
	}

	// The lock is only for serialising migrations for this isolated database.
	lockName := fmt.Sprintf("%s_tenant_%d", migrations.LockPrefix, tenantID)
	return migrations.RunUp(ctx, db, i.Dir, lockName)
}
