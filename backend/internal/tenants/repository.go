package tenants

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var ErrNotFound = errors.New("tenants: not found")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const selectCols = `
	id, name, slug, status,
	custom_domain, custom_domain_verified, custom_domain_token,
	created_at, updated_at, deleted_at
`

func (r *Repository) scan(row interface {
	Scan(dest ...any) error
}) (*Tenant, error) {
	var t Tenant
	if err := row.Scan(
		&t.ID, &t.Name, &t.Slug, &t.Status,
		&t.CustomDomain, &t.CustomDomainVerified, &t.CustomDomainToken,
		&t.CreatedAt, &t.UpdatedAt, &t.DeletedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

// ByCustomDomain resolves a tenant from a verified custom domain (Host
// header match). Only VERIFIED custom domains are matched — an attacker
// cannot claim someone else's traffic by pointing DNS at us before proving
// ownership.
func (r *Repository) ByCustomDomain(ctx context.Context, host string) (*Tenant, error) {
	q := fmt.Sprintf(`SELECT %s FROM tenants
		WHERE custom_domain = ? AND custom_domain_verified = 1
		AND status = 'active' AND deleted_at IS NULL LIMIT 1`, selectCols)
	row := r.db.QueryRowContext(ctx, q, strings.ToLower(strings.TrimSuffix(strings.TrimSpace(host), ".")))
	return r.scan(row)
}

// BySlug resolves a tenant from its slug, used for both subdomain
// (acme.gatepass.example.com) and path-prefix (gatepass.example.com/acme)
// resolution.
func (r *Repository) BySlug(ctx context.Context, slug string) (*Tenant, error) {
	q := fmt.Sprintf(`SELECT %s FROM tenants
		WHERE slug = ? AND status = 'active' AND deleted_at IS NULL LIMIT 1`, selectCols)
	row := r.db.QueryRowContext(ctx, q, strings.ToLower(strings.TrimSpace(slug)))
	return r.scan(row)
}

func (r *Repository) ByID(ctx context.Context, id int64) (*Tenant, error) {
	q := fmt.Sprintf(`SELECT %s FROM tenants WHERE id = ? AND deleted_at IS NULL LIMIT 1`, selectCols)
	row := r.db.QueryRowContext(ctx, q, id)
	return r.scan(row)
}

func (r *Repository) Create(ctx context.Context, t *Tenant) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO tenants (name, slug, status, custom_domain_token, created_at, updated_at)
		VALUES (?, ?, 'active', ?, NOW(), NOW())`,
		t.Name, t.Slug, t.CustomDomainToken,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// CreateTx is Create run inside an existing transaction — used by the
// platform bootstrap flow (tenant + first admin role/user/membership
// created atomically: either the whole tenant setup succeeds, or none of
// it exists).
func (r *Repository) CreateTx(ctx context.Context, tx *sql.Tx, t *Tenant) (int64, error) {
	res, err := tx.ExecContext(ctx, `
		INSERT INTO tenants (name, slug, status, custom_domain_token, created_at, updated_at)
		VALUES (?, ?, 'active', ?, NOW(), NOW())`,
		t.Name, t.Slug, t.CustomDomainToken,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// SetCustomDomain stores a pending (unverified) custom domain. It is not
// used for routing until VerifyCustomDomain succeeds.
func (r *Repository) SetCustomDomain(ctx context.Context, tenantID int64, domain string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE tenants SET custom_domain = ?, custom_domain_verified = 0, updated_at = NOW()
		WHERE id = ?`, domain, tenantID)
	return err
}

func (r *Repository) VerifyCustomDomain(ctx context.Context, tenantID int64) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE tenants SET custom_domain_verified = 1, updated_at = NOW()
		WHERE id = ? AND custom_domain IS NOT NULL`, tenantID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
