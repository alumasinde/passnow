package visitors

import (
	"context"
	"database/sql"
	"errors"
)

var ErrIDTypeNotFound = errors.New("visitors: id type not found")

type IDTypeRepository struct { db *sql.DB }

func NewIDTypeRepository(db *sql.DB) *IDTypeRepository { return &IDTypeRepository{db:db} }

func (r *IDTypeRepository) List(ctx context.Context, activeOnly bool) ([]IDType,error) {
	q:=`SELECT id,name,code,requires_number,active FROM id_types WHERE deleted_at IS NULL`
	if activeOnly { q+=" AND active = 1" }
	q+=" ORDER BY name"
	rows,err:=r.db.QueryContext(ctx,q);if err!=nil{return nil,err};defer rows.Close()
	var out []IDType
	for rows.Next(){var t IDType;if err:=rows.Scan(&t.ID,&t.Name,&t.Code,&t.RequiresNumber,&t.Active);err!=nil{return nil,err};out=append(out,t)}
	return out,rows.Err()
}

func (r *IDTypeRepository) ByID(ctx context.Context,id int64)(*IDType,error){
	var t IDType
	err:=r.db.QueryRowContext(ctx,`SELECT id,name,code,requires_number,active FROM id_types WHERE id=? AND deleted_at IS NULL LIMIT 1`,id).Scan(&t.ID,&t.Name,&t.Code,&t.RequiresNumber,&t.Active)
	if errors.Is(err,sql.ErrNoRows){return nil,ErrIDTypeNotFound};if err!=nil{return nil,err};return &t,nil
}

func (r *IDTypeRepository) Create(ctx context.Context,name,code string,requiresNumber,active bool)(int64,error){
	res,err:=r.db.ExecContext(ctx,`INSERT INTO id_types (name,code,requires_number,active,created_at,updated_at) VALUES (?,?,?,?,NOW(),NOW())`,name,code,requiresNumber,active);if err!=nil{return 0,err};return res.LastInsertId()
}

func (r *IDTypeRepository) Update(ctx context.Context,id int64,in IDTypeInput)error{
	t,err:=r.ByID(ctx,id);if err!=nil{return err}
	if in.Name!=""{t.Name=in.Name};if in.Code!=""{t.Code=in.Code};if in.RequiresNumber!=nil{t.RequiresNumber=*in.RequiresNumber};if in.Active!=nil{t.Active=*in.Active}
	_,err=r.db.ExecContext(ctx,`UPDATE id_types SET name=?,code=?,requires_number=?,active=?,updated_at=NOW() WHERE id=?`,t.Name,t.Code,t.RequiresNumber,t.Active,id);return err
}

func (r *IDTypeRepository) SoftDelete(ctx context.Context,id int64)error{
	res,err:=r.db.ExecContext(ctx,`UPDATE id_types SET deleted_at=NOW(),active=0 WHERE id=? AND deleted_at IS NULL`,id);if err!=nil{return err};n,_:=res.RowsAffected();if n==0{return ErrIDTypeNotFound};return nil
}
