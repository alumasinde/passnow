package gatedevices
import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)
var ErrNotFound=errors.New("gate device not found")
type Repository struct{db *sql.DB}
func NewRepository(db *sql.DB)*Repository{return &Repository{db}}
func (r *Repository) List(ctx context.Context)([]Device,error){rows,e:=r.db.QueryContext(ctx,"SELECT id,device_key,name,description,gate_id,active FROM gate_devices ORDER BY name");if e!=nil{return nil,e};defer rows.Close();var out []Device;for rows.Next(){var d Device;if e=rows.Scan(&d.ID,&d.DeviceKey,&d.Name,&d.Description,&d.GateID,&d.Active);e!=nil{return nil,e};out=append(out,d)};return out,rows.Err()}
func (r *Repository) Create(ctx context.Context,in Input)(*Device,error){a:=true;if in.Active!=nil{a=*in.Active};res,e:=r.db.ExecContext(ctx,"INSERT INTO gate_devices(device_key,device_secret_hash,name,description,gate_id,active) VALUES(?,?,?,?,?,?)",strings.TrimSpace(in.DeviceKey),secretHash(in.DeviceSecret),strings.TrimSpace(in.Name),in.Description,in.GateID,a);if e!=nil{return nil,e};id,_:=res.LastInsertId();return r.ByID(ctx,id)}
func(r *Repository)ByID(ctx context.Context,id int64)(*Device,error){d:=&Device{};e:=r.db.QueryRowContext(ctx,"SELECT id,device_key,name,description,gate_id,active FROM gate_devices WHERE id=?",id).Scan(&d.ID,&d.DeviceKey,&d.Name,&d.Description,&d.GateID,&d.Active);if errors.Is(e,sql.ErrNoRows){return nil,ErrNotFound};return d,e}

func (r *Repository) Resolve(ctx context.Context,key,secret string)(*Device,error){h:=sha256.Sum256([]byte(secret));d:=&Device{};e:=r.db.QueryRowContext(ctx,"SELECT id,device_key,name,description,gate_id,active,CAST(last_seen_at AS CHAR) FROM gate_devices WHERE device_key=? AND device_secret_hash=? AND active=1",key,fmt.Sprintf("%x",h)).Scan(&d.ID,&d.DeviceKey,&d.Name,&d.Description,&d.GateID,&d.Active,&d.LastSeenAt);if errors.Is(e,sql.ErrNoRows){return nil,ErrNotFound};if e==nil{_,_=r.db.ExecContext(ctx,"UPDATE gate_devices SET last_seen_at=NOW() WHERE id=?",d.ID)};return d,e}

func secretHash(secret string) string { h:=sha256.Sum256([]byte(secret)); return fmt.Sprintf("%x",h) }