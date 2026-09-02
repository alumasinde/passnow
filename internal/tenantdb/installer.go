package tenantdb

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gatepass/internal/migrations"
	_ "github.com/go-sql-driver/mysql"
)

type Installer struct{ Dir string }
func NewInstaller(dir string)*Installer{if dir==""{dir="migrations/tenant"};return &Installer{Dir:dir}}

func (i *Installer) RunUp(ctx context.Context, creds Credentials, tenantID int64) error {
	if tenantID<1{return fmt.Errorf("tenantdb: invalid tenant id")}
	db,err:=sql.Open("mysql",mysqlDSN(creds));if err!=nil{return err};defer db.Close()
	pctx,cancel:=context.WithTimeout(ctx,10*time.Second);defer cancel()
	if err:=db.PingContext(pctx);err!=nil{return err}
	lockName:=fmt.Sprintf("%s_tenant_%d",migrations.LockPrefix,tenantID)
	return migrations.RunUp(ctx,db,i.Dir,lockName)
}

func (i *Installer) Install(ctx context.Context, creds Credentials, tenantID int64, tenantName, slug, domainToken string) error {
	return i.RunUp(ctx,creds,tenantID)
}
