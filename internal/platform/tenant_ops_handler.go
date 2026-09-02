package platform

import (
 "context"
 "database/sql"
 "errors"
 "fmt"
 "net/http"
 "strconv"
 "time"

 "gatepass/internal/httpx"
 "gatepass/internal/middleware"
 "gatepass/internal/tenantdb"
)

type TenantOpsHandler struct{ manager *tenantdb.Manager; installer *tenantdb.Installer }

func NewTenantOpsHandler(manager *tenantdb.Manager, installer *tenantdb.Installer)*TenantOpsHandler{return &TenantOpsHandler{manager:manager,installer:installer}}

func (h *TenantOpsHandler) Health(w http.ResponseWriter,r *http.Request){
 id,ok:=opsTenantID(w,r);if !ok{return}
 ctx,cancel:=context.WithTimeout(r.Context(),8*time.Second);defer cancel()
 creds,err:=h.manager.Credentials(ctx,id)
 if err!=nil{httpx.WriteJSON(w,http.StatusOK,map[string]any{"tenant_id":id,"healthy":false,"status":"unavailable","message":"Tenant database credentials are unavailable"});return}
 db,err:=h.manager.DB(ctx,id)
 if err==nil {err=db.PingContext(ctx)}
 healthy:=err==nil
 message:="Database connection is healthy"
 if !healthy {message="Database health check failed"}
 httpx.WriteJSON(w,http.StatusOK,map[string]any{"tenant_id":id,"healthy":healthy,"status":map[bool]string{true:"healthy",false:"unhealthy"}[healthy],"database":creds.Database,"message":message})
}

func (h *TenantOpsHandler) Migrate(w http.ResponseWriter,r *http.Request){
 id,ok:=opsTenantID(w,r);if !ok{return}
 ctx,cancel:=context.WithTimeout(r.Context(),60*time.Second);defer cancel()
 creds,err:=h.manager.Credentials(ctx,id)
 if err!=nil{httpx.WriteError(w,httpx.ErrValidation.WithMessage("tenant database is not ready"));return}
 // Invalidate cached pool after schema work so the next tenant request uses a fresh connection.
 if err:=h.installer.RunUp(ctx,creds,id);err!=nil{httpx.WriteError(w,httpx.ErrValidation.WithMessage(fmt.Sprintf("migration failed: %v",err)));return}
 h.manager.Invalidate(id)
 httpx.WriteJSON(w,http.StatusOK,map[string]any{"tenant_id":id,"status":"ok","message":"Tenant migrations completed successfully"})
}

func opsTenantID(w http.ResponseWriter,r *http.Request)(int64,bool){id,err:=strconv.ParseInt(r.PathValue("id"),10,64);if err!=nil||id<1{httpx.WriteError(w,httpx.ErrValidation.WithMessage("invalid tenant id"));return 0,false};return id,true}

// Compile-time check keeps platform routes explicitly protected at registration sites.
var _ func([]byte,*AdminRepository,http.Handler)http.Handler = middleware.PlatformAdmin
var _ = errors.New
var _ *sql.DB
