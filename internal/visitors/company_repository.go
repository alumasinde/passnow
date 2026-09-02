package visitors

import (
	"context"
	"database/sql"
	"errors"

	"gatepass/internal/database"
	"gatepass/internal/httpx"
)

var (
	ErrCompanyNotFound=errors.New("visitors: company not found")
	ErrCompanyNameTaken=errors.New("visitors: company name already exists")
)

type CompanyRepository struct{db *sql.DB}
func NewCompanyRepository(db *sql.DB)*CompanyRepository{return &CompanyRepository{db:db}}

func (r *CompanyRepository) List(ctx context.Context,activeOnly bool,p httpx.Pagination)([]Company,int,error){
	where:="WHERE deleted_at IS NULL"
	if activeOnly{where+=" AND active=1"}
	var total int;if err:=r.db.QueryRowContext(ctx,"SELECT COUNT(*) FROM visitor_companies "+where).Scan(&total);err!=nil{return nil,0,err}
	rows,err:=r.db.QueryContext(ctx,"SELECT id,name,phone,email,address,active FROM visitor_companies "+where+" ORDER BY name LIMIT ? OFFSET ?",p.Limit,p.Offset);if err!=nil{return nil,0,err};defer rows.Close()
	var out []Company;for rows.Next(){var c Company;if err:=rows.Scan(&c.ID,&c.Name,&c.Phone,&c.Email,&c.Address,&c.Active);err!=nil{return nil,0,err};out=append(out,c)};return out,total,rows.Err()
}

func (r *CompanyRepository) ByID(ctx context.Context,id int64)(*Company,error){
	var c Company;err:=r.db.QueryRowContext(ctx,`SELECT id,name,phone,email,address,active FROM visitor_companies WHERE id=? AND deleted_at IS NULL LIMIT 1`,id).Scan(&c.ID,&c.Name,&c.Phone,&c.Email,&c.Address,&c.Active)
	if errors.Is(err,sql.ErrNoRows){return nil,ErrCompanyNotFound};if err!=nil{return nil,err};return &c,nil
}

func (r *CompanyRepository) Create(ctx context.Context,in CompanyInput)(int64,error){
	res,err:=r.db.ExecContext(ctx,`INSERT INTO visitor_companies (name,phone,email,address,active,created_at,updated_at) VALUES (?,?,?,?,1,NOW(),NOW())`,in.Name,in.Phone,in.Email,in.Address)
	if err!=nil{if database.IsDuplicateKeyErr(err){return 0,ErrCompanyNameTaken};return 0,err};return res.LastInsertId()
}

func (r *CompanyRepository) Update(ctx context.Context,id int64,in CompanyInput)error{
	c,err:=r.ByID(ctx,id);if err!=nil{return err}
	if in.Name!=""{c.Name=in.Name};if in.Phone!=nil{c.Phone=in.Phone};if in.Email!=nil{c.Email=in.Email};if in.Address!=nil{c.Address=in.Address};if in.Active!=nil{c.Active=*in.Active}
	_,err=r.db.ExecContext(ctx,`UPDATE visitor_companies SET name=?,phone=?,email=?,address=?,active=?,updated_at=NOW() WHERE id=?`,c.Name,c.Phone,c.Email,c.Address,c.Active,id)
	if err!=nil&&database.IsDuplicateKeyErr(err){return ErrCompanyNameTaken};return err
}
