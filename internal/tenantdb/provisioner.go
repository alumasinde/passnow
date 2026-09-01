package tenantdb

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var databaseIdentifier = regexp.MustCompile("^[A-Za-z0-9_]{1,64}$")

type Provisioner struct {
	host, port, username, password string
}

func NewProvisioner(host, port, username, password string) *Provisioner {
	return &Provisioner{host: host, port: port, username: username, password: password}
}

func (p *Provisioner) Host() string { return p.host }

func (p *Provisioner) Port() string { return p.port }

// Enabled reports whether provisioning can connect. An empty password is valid
// for local MySQL/MariaDB installations, so only connection location and user
// are required here.
func (p *Provisioner) Enabled() bool {
	return p.host != "" && p.port != "" && p.username != ""
}

func (p *Provisioner) CreateDatabase(ctx context.Context, name string) error {
	if !p.Enabled() { return fmt.Errorf("tenantdb: provisioning database credentials are not configured") }
	if !databaseIdentifier.MatchString(name) { return fmt.Errorf("tenantdb: invalid database name") }
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci&loc=UTC", p.username, p.password, p.host, p.port)
	db, err := sql.Open("mysql", dsn)
	if err != nil { return err }
	defer db.Close()
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil { return err }
	_, err = db.ExecContext(ctx, "CREATE DATABASE IF NOT EXISTS `"+name+"` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci")
	return err
}

func Verify(ctx context.Context, c Credentials) error {
	if c.Host == "" || c.Port == "" || c.Database == "" || c.Username == "" {
		return fmt.Errorf("tenantdb: host, port, database, and username are required")
	}
	db, err := sql.Open("mysql", mysqlDSN(c))
	if err != nil { return err }
	defer db.Close()
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return db.PingContext(pingCtx)
}
