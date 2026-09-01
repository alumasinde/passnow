package tenants

import (
 "context"
 "strings"
 "fmt"
 "database/sql"
 "time"
)

type DomainType string
const (DomainSubdomain DomainType="subdomain"; DomainCustom DomainType="custom")
type Domain struct { ID int64; TenantID int64; Domain string; Type DomainType; IsPrimary bool; IsVerified bool; CreatedAt time.Time; UpdatedAt time.Time }

func normalizeDomain(v string) string { return strings.ToLower(strings.TrimSuffix(strings.TrimSpace(v),".")) }

func (r *Repository) Domains(ctx context.Context, tenantID int64)([]Domain,error){
 rows,err:=r.db.QueryContext(ctx,`SELECT id,tenant_id,domain,domain_type,is_primary,is_verified,created_at,updated_at FROM tenant_domains WHERE tenant_id=? ORDER BY is_primary DESC,id ASC`,tenantID);if err!=nil{return nil,err};defer rows.Close()
 out:=[]Domain{};for rows.Next(){var d Domain;if err:=rows.Scan(&d.ID,&d.TenantID,&d.Domain,&d.Type,&d.IsPrimary,&d.IsVerified,&d.CreatedAt,&d.UpdatedAt);err!=nil{return nil,err};out=append(out,d)};return out,rows.Err()
}
func (r *Repository) AddDomain(ctx context.Context, tenantID int64, domain string, typ DomainType, primary, verified bool) error {
 domain=normalizeDomain(domain); if domain=="" { return ErrNotFound }
 tx,err:=r.db.BeginTx(ctx,nil);if err!=nil{return err};defer tx.Rollback()
 if primary { if _,err=tx.ExecContext(ctx,`UPDATE tenant_domains SET is_primary=0 WHERE tenant_id=?`,tenantID);err!=nil{return err} }
 _,err=tx.ExecContext(ctx,`INSERT INTO tenant_domains (tenant_id,domain,domain_type,is_primary,is_verified) VALUES (?,?,?,?,?)`,tenantID,domain,typ,primary,verified);if err!=nil{return err}
 return tx.Commit()
}
func (r *Repository) DeleteDomain(ctx context.Context, tenantID, domainID int64) error {_,err:=r.db.ExecContext(ctx,`DELETE FROM tenant_domains WHERE id=? AND tenant_id=? AND domain_type='custom'`,domainID,tenantID);return err}
func (r *Repository) SetPrimaryDomain(ctx context.Context, tenantID, domainID int64) error {tx,err:=r.db.BeginTx(ctx,nil);if err!=nil{return err};defer tx.Rollback();if _,err=tx.ExecContext(ctx,`UPDATE tenant_domains SET is_primary=0 WHERE tenant_id=?`,tenantID);err!=nil{return err};res,err:=tx.ExecContext(ctx,`UPDATE tenant_domains SET is_primary=1 WHERE id=? AND tenant_id=?`,domainID,tenantID);if err!=nil{return err};n,_:=res.RowsAffected();if n==0{return ErrNotFound};return tx.Commit()}
func (r *Repository) ByDomain(ctx context.Context, host string)(*Tenant,error){
 host=normalizeDomain(host);q:=fmt.Sprintf("SELECT %s FROM tenants t INNER JOIN tenant_domains d ON d.tenant_id=t.id WHERE d.domain=? AND d.is_verified=1 AND t.status='active' AND t.deleted_at IS NULL LIMIT 1",selectCols)
 return r.scan(r.db.QueryRowContext(ctx,q,host))
}


// ReconcilePlatformDomains ensures every existing tenant has exactly one
// system-owned subdomain derived from its current slug. It also repairs older
// installations that stored *.BASE_DOMAIN entries as custom domains.
func (r *Repository) ReconcilePlatformDomains(ctx context.Context, baseDomain string) error {
 baseDomain = normalizeDomain(baseDomain)
 if baseDomain == "" || baseDomain == "localhost" { return nil }

 rows, err := r.db.QueryContext(ctx, `SELECT id, slug FROM tenants WHERE deleted_at IS NULL`)
 if err != nil { return err }
 defer rows.Close()
 type tenantRow struct{ id int64; slug string }
 var items []tenantRow
 for rows.Next() {
  var item tenantRow
  if err := rows.Scan(&item.id, &item.slug); err != nil { return err }
  items = append(items, item)
 }
 if err := rows.Err(); err != nil { return err }

 for _, item := range items {
  generated := normalizeDomain(item.slug + "." + baseDomain)
  tx, err := r.db.BeginTx(ctx, nil)
  if err != nil { return err }

  // If an old platform-domain row exists under BASE_DOMAIN, convert it into
  // the canonical slug-based system subdomain. This fixes older rows that
  // were incorrectly saved as custom/pending domains.
  var oldID int64
  err = tx.QueryRowContext(ctx, `
   SELECT id FROM tenant_domains
   WHERE tenant_id=? AND domain LIKE CONCAT('%.', ?) AND domain_type='custom'
   ORDER BY is_primary DESC, id ASC LIMIT 1`, item.id, baseDomain).Scan(&oldID)
  if err != nil && err != sql.ErrNoRows {
   _ = tx.Rollback()
   return err
  }

  var exists int
  err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM tenant_domains WHERE tenant_id=? AND domain=?`, item.id, generated).Scan(&exists)
  if err != nil {
   _ = tx.Rollback()
   return err
  }
  if oldID != 0 && exists == 0 {
   _, err = tx.ExecContext(ctx, `
    UPDATE tenant_domains
    SET domain=?, domain_type='subdomain', is_verified=1, updated_at=NOW()
    WHERE id=? AND tenant_id=?`, generated, oldID, item.id)
   if err != nil {
    _ = tx.Rollback()
    return err
   }
   exists = 1
  }
  if err != nil {
   _ = tx.Rollback()
   return err
  }
  if exists == 0 {
   var count int
   if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM tenant_domains WHERE tenant_id=?`, item.id).Scan(&count); err != nil {
    _ = tx.Rollback()
    return err
   }
   _, err = tx.ExecContext(ctx, `
    INSERT INTO tenant_domains (tenant_id,domain,domain_type,is_primary,is_verified)
    VALUES (?,?, 'subdomain', ?, 1)`, item.id, generated, count == 0)
   if err != nil {
    _ = tx.Rollback()
    return err
   }
  }

  if err := tx.Commit(); err != nil { return err }
 }
 return nil
}
