package tenants

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "strings"
)

var ErrNotFound = errors.New("tenants: not found")

type Repository struct { db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

const selectCols = `id, name, slug, status, custom_domain, custom_domain_verified, custom_domain_token, created_at, updated_at, deleted_at`

func (r *Repository) scan(row interface{ Scan(dest ...any) error }) (*Tenant, error) {
    var t Tenant
    if err := row.Scan(&t.ID,&t.Name,&t.Slug,&t.Status,&t.CustomDomain,&t.CustomDomainVerified,&t.CustomDomainToken,&t.CreatedAt,&t.UpdatedAt,&t.DeletedAt); err != nil {
        if errors.Is(err, sql.ErrNoRows) { return nil, ErrNotFound }
        return nil, err
    }
    return &t, nil
}

func (r *Repository) ByCustomDomain(ctx context.Context, host string) (*Tenant,error) {
    q:=fmt.Sprintf("SELECT %s FROM tenants WHERE custom_domain=? AND custom_domain_verified=1 AND status='active' AND deleted_at IS NULL LIMIT 1",selectCols)
    return r.scan(r.db.QueryRowContext(ctx,q,strings.ToLower(strings.TrimSuffix(strings.TrimSpace(host),"."))))
}
func (r *Repository) BySlug(ctx context.Context, slug string) (*Tenant,error) {
    q:=fmt.Sprintf("SELECT %s FROM tenants WHERE slug=? AND status='active' AND deleted_at IS NULL LIMIT 1",selectCols)
    return r.scan(r.db.QueryRowContext(ctx,q,strings.ToLower(strings.TrimSpace(slug))))
}
func (r *Repository) ByID(ctx context.Context,id int64)(*Tenant,error){
    q:=fmt.Sprintf("SELECT %s FROM tenants WHERE id=? AND deleted_at IS NULL LIMIT 1",selectCols)
    return r.scan(r.db.QueryRowContext(ctx,q,id))
}
func (r *Repository) List(ctx context.Context)([]*Tenant,error){
    q:=fmt.Sprintf("SELECT %s FROM tenants WHERE deleted_at IS NULL ORDER BY created_at DESC",selectCols)
    rows,err:=r.db.QueryContext(ctx,q); if err!=nil{return nil,err}; defer rows.Close()
    out:=make([]*Tenant,0)
    for rows.Next(){ t,err:=r.scan(rows); if err!=nil{return nil,err}; out=append(out,t) }
    return out,rows.Err()
}
func (r *Repository) Create(ctx context.Context,t *Tenant)(int64,error){
    res,err:=r.db.ExecContext(ctx,"INSERT INTO tenants (name,slug,status,custom_domain_token,created_at,updated_at) VALUES (?,?, 'active', ?,NOW(),NOW())",t.Name,t.Slug,t.CustomDomainToken);if err!=nil{return 0,err};return res.LastInsertId()
}
func (r *Repository) CreateTx(ctx context.Context,tx *sql.Tx,t *Tenant)(int64,error){
    res,err:=tx.ExecContext(ctx,"INSERT INTO tenants (name,slug,status,custom_domain_token,created_at,updated_at) VALUES (?,?, 'active', ?,NOW(),NOW())",t.Name,t.Slug,t.CustomDomainToken);if err!=nil{return 0,err};return res.LastInsertId()
}
func (r *Repository) UpdateStatus(ctx context.Context,id int64,status Status) error {
    res,err:=r.db.ExecContext(ctx,"UPDATE tenants SET status=?, updated_at=NOW() WHERE id=? AND deleted_at IS NULL",status,id);if err!=nil{return err};n,_:=res.RowsAffected();if n==0{return ErrNotFound};return nil
}
func (r *Repository) SetCustomDomain(ctx context.Context,id int64,domain string)error{_,err:=r.db.ExecContext(ctx,"UPDATE tenants SET custom_domain=?, custom_domain_verified=0,updated_at=NOW() WHERE id=?",domain,id);return err}
func (r *Repository) VerifyCustomDomain(ctx context.Context,id int64)error{res,err:=r.db.ExecContext(ctx,"UPDATE tenants SET custom_domain_verified=1,updated_at=NOW() WHERE id=? AND custom_domain IS NOT NULL",id);if err!=nil{return err};n,_:=res.RowsAffected();if n==0{return ErrNotFound};return nil}


func (r *Repository) Update(ctx context.Context, id int64, name, slug string) error {
 res,err:=r.db.ExecContext(ctx,"UPDATE tenants SET name=?, slug=?, updated_at=NOW() WHERE id=? AND deleted_at IS NULL",strings.TrimSpace(name),strings.ToLower(strings.TrimSpace(slug)),id);if err!=nil{return err};n,_:=res.RowsAffected();if n==0{return ErrNotFound};return nil
}

// ReadyIDs returns active tenants whose provisioned database is ready.
// It keeps operational workers from repeatedly attempting tenants that are
// still provisioning or whose database verification failed.
func (r *Repository) ReadyIDs(ctx context.Context) ([]int64, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.id
		FROM tenants t
		INNER JOIN tenant_databases td ON td.tenant_id = t.id
		WHERE t.status = 'active'
		  AND t.deleted_at IS NULL
		  AND td.status = 'ready'
		ORDER BY t.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
