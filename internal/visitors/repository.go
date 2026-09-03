package visitors

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"gatepass/internal/database"
	"gatepass/internal/httpx"
)

var (
	ErrNotFound=errors.New("visitors: not found")
	ErrDuplicateIDNumber=errors.New("visitors: this ID document is already registered")
)

type Repository struct{db *sql.DB}
func NewRepository(db *sql.DB)*Repository{return &Repository{db:db}}

const selectCols=`id,first_name,last_name,id_type_id,id_number,company_id,phone,email,photo_ref,notes,source,status,blacklist_reason,created_by,updated_by,created_at,updated_at,deleted_at`

func (r *Repository) scan(row interface{Scan(dest ...any)error})(*Visitor,error){
	var v Visitor
	if err:=row.Scan(&v.ID,&v.FirstName,&v.LastName,&v.IDTypeID,&v.IDNumber,&v.CompanyID,&v.Phone,&v.Email,&v.PhotoRef,&v.Notes,&v.Source,&v.Status,&v.BlacklistReason,&v.CreatedBy,&v.UpdatedBy,&v.CreatedAt,&v.UpdatedAt,&v.DeletedAt);err!=nil{if errors.Is(err,sql.ErrNoRows){return nil,ErrNotFound};return nil,err};return &v,nil
}

func (r *Repository) ByID(ctx context.Context,id int64)(*Visitor,error){
	return r.scan(r.db.QueryRowContext(ctx,"SELECT "+selectCols+" FROM visitors WHERE id=? AND deleted_at IS NULL LIMIT 1",id))
}

type ListFilter struct{Status *Status;CompanyID *int64;Blacklisted *bool;Search string;CreatedBy *int64}

func (r *Repository) List(ctx context.Context,f ListFilter,p httpx.Pagination)([]Visitor,int,error){
	where:="WHERE deleted_at IS NULL";args:=[]any{}
	if f.Status!=nil{where+=" AND status=?";args=append(args,*f.Status)}
	if f.CompanyID!=nil{where+=" AND company_id=?";args=append(args,*f.CompanyID)}
	if f.CreatedBy!=nil{where+=" AND created_by=?";args=append(args,*f.CreatedBy)}
	if f.Blacklisted!=nil{if *f.Blacklisted{where+=" AND status=?";args=append(args,StatusBlacklisted)}else{where+=" AND status<>?";args=append(args,StatusBlacklisted)}}
	if f.Search!=""{where+=" AND (id_number=? OR first_name LIKE ? OR last_name LIKE ? OR phone LIKE ? OR email LIKE ?)";like:=f.Search+"%";args=append(args,f.Search,like,like,like,like)}
	var total int;if err:=r.db.QueryRowContext(ctx,"SELECT COUNT(*) FROM visitors "+where,args...).Scan(&total);err!=nil{return nil,0,err}
	orderBy:="ORDER BY created_at DESC";selectArgs:=append([]any{},args...)
	if f.Search!=""{orderBy="ORDER BY (id_number = ?) DESC, created_at DESC";selectArgs=append(selectArgs,f.Search)}
	selectArgs=append(selectArgs,p.Limit,p.Offset)
	rows,err:=r.db.QueryContext(ctx,"SELECT "+selectCols+" FROM visitors "+where+" "+orderBy+" LIMIT ? OFFSET ?",selectArgs...);if err!=nil{return nil,0,err};defer rows.Close()
	var out []Visitor;for rows.Next(){v,err:=r.scan(rows);if err!=nil{return nil,0,err};out=append(out,*v)};return out,total,rows.Err()
}

func (r *Repository) Create(ctx context.Context,v *Visitor)(int64,error){
	res,err:=r.db.ExecContext(ctx,`INSERT INTO visitors (first_name,last_name,id_type_id,id_number,company_id,phone,email,photo_ref,notes,source,status,created_by,updated_by,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,'active',?,?,NOW(),NOW())`,v.FirstName,v.LastName,v.IDTypeID,v.IDNumber,v.CompanyID,v.Phone,v.Email,v.PhotoRef,v.Notes,v.Source,v.CreatedBy,v.CreatedBy)
	if err!=nil{if database.IsDuplicateKeyErr(err){return 0,ErrDuplicateIDNumber};return 0,err};return res.LastInsertId()
}

func (r *Repository) Update(ctx context.Context,id int64,in UpdateInput,updatedBy int64)(*Visitor,error){
	v,err:=r.ByID(ctx,id);if err!=nil{return nil,err}
	if in.FirstName!=nil{v.FirstName=*in.FirstName};if in.LastName!=nil{v.LastName=*in.LastName};if in.IDTypeID!=nil{v.IDTypeID=*in.IDTypeID};if in.IDNumber!=nil{v.IDNumber=in.IDNumber};if in.CompanyID!=nil{v.CompanyID=in.CompanyID};if in.Phone!=nil{v.Phone=in.Phone};if in.Email!=nil{v.Email=in.Email};if in.Notes!=nil{v.Notes=in.Notes}
	_,err=r.db.ExecContext(ctx,`UPDATE visitors SET first_name=?,last_name=?,id_type_id=?,id_number=?,company_id=?,phone=?,email=?,notes=?,updated_by=?,updated_at=NOW() WHERE id=?`,v.FirstName,v.LastName,v.IDTypeID,v.IDNumber,v.CompanyID,v.Phone,v.Email,v.Notes,updatedBy,id)
	if err!=nil{if database.IsDuplicateKeyErr(err){return nil,ErrDuplicateIDNumber};return nil,err};v.UpdatedBy=&updatedBy;return v,nil
}

func (r *Repository) SetBlacklist(ctx context.Context,id int64,blacklisted bool,reason *string,updatedBy int64)error{
	status:=StatusActive;if blacklisted{status=StatusBlacklisted}else{reason=nil}
	res,err:=r.db.ExecContext(ctx,`UPDATE visitors SET status=?,blacklist_reason=?,updated_by=?,updated_at=NOW() WHERE id=? AND deleted_at IS NULL`,status,reason,updatedBy,id);if err!=nil{return err};n,_:=res.RowsAffected();if n==0{return ErrNotFound};return nil
}


func (r *Repository) IdentityMatches(ctx context.Context, idTypeID int64, idNumber, phone, email string, limit int) ([]Visitor,error) {
 if limit<1||limit>20 { limit=10 }; where:="WHERE deleted_at IS NULL AND ("; args:=[]any{}; parts:=[]string{}
 if idNumber!="" { parts=append(parts,"(id_type_id=? AND id_number=?)"); args=append(args,idTypeID,idNumber) }
 if phone!="" { parts=append(parts,"phone=?"); args=append(args,phone) }
 if email!="" { parts=append(parts,"LOWER(email)=LOWER(?)"); args=append(args,email) }
 if len(parts)==0 { return []Visitor{},nil }; where+=strings.Join(parts," OR ")+")"; args=append(args,limit)
 rows,err:=r.db.QueryContext(ctx,"SELECT "+selectCols+" FROM visitors "+where+" ORDER BY created_at DESC LIMIT ?",args...);if err!=nil{return nil,err};defer rows.Close();out:=[]Visitor{};for rows.Next(){v,e:=r.scan(rows);if e!=nil{return nil,e};out=append(out,*v)};return out,rows.Err()
}
