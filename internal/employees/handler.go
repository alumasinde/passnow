package employees

import (
	"context"
	"errors"
	"gatepass/internal/rbac"
	"net/http"
	"strconv"

	"gatepass/internal/httpx"
	"gatepass/internal/reqctx"
)

type Handler struct { svc *Service }
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }
func requireTenant(w http.ResponseWriter, r *http.Request) bool { if _,ok:=reqctx.TenantFromContext(r.Context()); !ok { httpx.WriteError(w,httpx.ErrAuthRequired); return false }; return true }

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	if !requireTenant(w,r){return}
	p:=httpx.ParsePagination(r)
	var departmentID *int64
	if d,ok:=rbac.DecisionFromContext(r.Context());ok&&d.Scope==rbac.ScopeDepartment{
		claims,ok:=reqctx.ClaimsFromContext(r.Context());if !ok{httpx.WriteError(w,httpx.ErrAuthRequired);return}
		var err error;departmentID,err=h.svc.UserDepartment(r.Context(),claims.UserID);if err!=nil||departmentID==nil{httpx.WriteError(w,httpx.ErrForbidden);return}
	}
	items,total,err:=h.svc.ListScoped(r.Context(),p,departmentID);if err!=nil{httpx.WriteError(w,httpx.ErrInternal);return}
	dtos:=make([]DTO,0,len(items));for i:=range items{dtos=append(dtos,h.svc.ToDTO(r.Context(),&items[i]))}
	httpx.WriteJSON(w,http.StatusOK,httpx.ListEnvelope[DTO]{Items:dtos,Limit:p.Limit,Offset:p.Offset,Total:total})
}

func (h *Handler) Get(w http.ResponseWriter,r *http.Request){if !requireTenant(w,r){return};id,err:=strconv.ParseInt(r.PathValue("id"),10,64);if err!=nil||id<1{httpx.WriteError(w,httpx.ErrNotFound);return};e,err:=h.svc.Get(r.Context(),id);if err!=nil{httpx.WriteError(w,httpx.ErrNotFound);return};httpx.WriteJSON(w,http.StatusOK,h.svc.ToDTO(r.Context(),e))}
