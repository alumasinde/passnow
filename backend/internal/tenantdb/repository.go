package tenantdb

import (
	"context"
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("tenantdb: connection not found")

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Get(ctx context.Context, tenantID int64) (*Connection, error) {
	c := &Connection{}
	err := r.db.QueryRowContext(ctx, "SELECT tenant_id, host, port, database_name, username, encrypted_password, status, verified_at, last_error, created_at, updated_at FROM tenant_databases WHERE tenant_id = ? LIMIT 1", tenantID).
		Scan(&c.TenantID, &c.Host, &c.Port, &c.DatabaseName, &c.Username, &c.EncryptedPassword, &c.Status, &c.VerifiedAt, &c.LastError, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) { return nil, ErrNotFound }
	if err != nil { return nil, err }
	return c, nil
}

func (r *Repository) Upsert(ctx context.Context, c *Connection) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO tenant_databases (tenant_id, host, port, database_name, username, encrypted_password, status, verified_at, last_error) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE host=VALUES(host), port=VALUES(port), database_name=VALUES(database_name), username=VALUES(username), encrypted_password=VALUES(encrypted_password), status=VALUES(status), verified_at=VALUES(verified_at), last_error=VALUES(last_error), updated_at=NOW()",
		c.TenantID, c.Host, c.Port, c.DatabaseName, c.Username, c.EncryptedPassword, c.Status, c.VerifiedAt, c.LastError)
	return err
}

func (r *Repository) MarkStatus(ctx context.Context, tenantID int64, status Status, verified bool, lastErr *string) error {
	if verified {
		_, err := r.db.ExecContext(ctx, "UPDATE tenant_databases SET status=?, verified_at=NOW(), last_error=NULL, updated_at=NOW() WHERE tenant_id=?", status, tenantID)
		return err
	}
	_, err := r.db.ExecContext(ctx, "UPDATE tenant_databases SET status=?, last_error=?, updated_at=NOW() WHERE tenant_id=?", status, lastErr, tenantID)
	return err
}
