package gates

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

var (
	ErrNotFound = errors.New("gates: not found")
	ErrDuplicateCode = errors.New("gates: code already exists")
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

const columns = `id, code, name, description, location, allows_entry, allows_exit, is_default, active, created_at, updated_at`

func scan(row interface{ Scan(...any) error }) (*Gate, error) {
	var g Gate
	if err := row.Scan(&g.ID,&g.Code,&g.Name,&g.Description,&g.Location,&g.AllowsEntry,&g.AllowsExit,&g.IsDefault,&g.Active,&g.CreatedAt,&g.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) { return nil, ErrNotFound }
		return nil, err
	}
	return &g,nil
}

func (r *Repository) List(ctx context.Context, all bool) ([]Gate,error) {
	q := "SELECT "+columns+" FROM gates WHERE deleted_at IS NULL"
	if !all { q += " AND active=1" }
	q += " ORDER BY is_default DESC, name ASC"
	rows,err:=r.db.QueryContext(ctx,q); if err!=nil{return nil,err}; defer rows.Close()
	out:=[]Gate{}
	for rows.Next(){g,err:=scan(rows);if err!=nil{return nil,err};out=append(out,*g)}
	return out,rows.Err()
}

func (r *Repository) ByID(ctx context.Context,id int64)(*Gate,error){
	return scan(r.db.QueryRowContext(ctx,"SELECT "+columns+" FROM gates WHERE id=? AND deleted_at IS NULL",id))
}

func clean(v *string)*string { if v==nil{return nil}; s:=strings.TrimSpace(*v); if s==""{return nil}; return &s }

func (r *Repository) Create(ctx context.Context,in Input)(*Gate,error){
	entry,exit,def,active:=true,true,false,true
	if in.AllowsEntry!=nil{entry=*in.AllowsEntry}; if in.AllowsExit!=nil{exit=*in.AllowsExit}; if in.IsDefault!=nil{def=*in.IsDefault}; if in.Active!=nil{active=*in.Active}
	if !entry && !exit{return nil,errors.New("a gate must allow entry, exit, or both")}
	tx,err:=r.db.BeginTx(ctx,nil);if err!=nil{return nil,err};defer tx.Rollback()
	if def { if _,err=tx.ExecContext(ctx,"UPDATE gates SET is_default=0 WHERE deleted_at IS NULL");err!=nil{return nil,err} }
	res,err:=tx.ExecContext(ctx,`INSERT INTO gates (code,name,description,location,allows_entry,allows_exit,is_default,active) VALUES (?,?,?,?,?,?,?,?)`,strings.ToUpper(strings.TrimSpace(in.Code)),strings.TrimSpace(in.Name),clean(in.Description),clean(in.Location),entry,exit,def,active)
	if err!=nil{return nil,err}; id,err:=res.LastInsertId();if err!=nil{return nil,err};if err=tx.Commit();err!=nil{return nil,err};return r.ByID(ctx,id)
}

func (r *Repository) Update(ctx context.Context,id int64,in Input)(*Gate,error){
	current,err:=r.ByID(ctx,id);if err!=nil{return nil,err}
	code,name:=current.Code,current.Name;if strings.TrimSpace(in.Code)!=""{code=strings.ToUpper(strings.TrimSpace(in.Code))};if strings.TrimSpace(in.Name)!=""{name=strings.TrimSpace(in.Name)}
	entry,exit,def,active:=current.AllowsEntry,current.AllowsExit,current.IsDefault,current.Active
	if in.AllowsEntry!=nil{entry=*in.AllowsEntry};if in.AllowsExit!=nil{exit=*in.AllowsExit};if in.IsDefault!=nil{def=*in.IsDefault};if in.Active!=nil{active=*in.Active}
	if !entry&&!exit{return nil,errors.New("a gate must allow entry, exit, or both")}
	tx,err:=r.db.BeginTx(ctx,nil);if err!=nil{return nil,err};defer tx.Rollback()
	if def {if _,err=tx.ExecContext(ctx,"UPDATE gates SET is_default=0 WHERE id<>? AND deleted_at IS NULL",id);err!=nil{return nil,err}}
	if _,err=tx.ExecContext(ctx,`UPDATE gates SET code=?,name=?,description=?,location=?,allows_entry=?,allows_exit=?,is_default=?,active=?,updated_at=NOW() WHERE id=? AND deleted_at IS NULL`,code,name,clean(in.Description),clean(in.Location),entry,exit,def,active,id);err!=nil{return nil,err}
	if err=tx.Commit();err!=nil{return nil,err};return r.ByID(ctx,id)
}