package platform

import (
 "net/http"
 "strconv"
 "strings"
 "unicode"

 "gatepass/internal/httpx"
 "gatepass/internal/tenants"
)

type TenantHandler struct{ repo *tenants.Repository }
func NewTenantHandler(repo *tenants.Repository)*TenantHandler{return &TenantHandler{repo:repo}}

func tenantDTO(t *tenants.Tenant) map[string]any{return map[string]any{"id":t.ID,"name":t.Name,"slug":t.Slug,"status":t.Status,"created_at":t.CreatedAt,"updated_at":t.UpdatedAt}}

func (h *TenantHandler) List(w http.ResponseWriter,r *http.Request){items,err:=h.repo.List(r.Context());if err!=nil{httpx.WriteError(w,httpx.ErrInternal);return};out:=make([]map[string]any,0,len(items));for _,t:=range items{out=append(out,tenantDTO(t))};httpx.WriteJSON(w,http.StatusOK,map[string]any{"tenants":out})}
func (h *TenantHandler) Get(w http.ResponseWriter,r *http.Request){id,ok:=platformID(w,r);if !ok{return};t,err:=h.repo.ByID(r.Context(),id);if err==tenants.ErrNotFound{httpx.WriteError(w,httpx.ErrNotFound);return};if err!=nil{httpx.WriteError(w,httpx.ErrInternal);return};httpx.WriteJSON(w,http.StatusOK,tenantDTO(t))}
type tenantStatusRequest struct{Status string `json:"status"`}
func (h *TenantHandler) UpdateStatus(w http.ResponseWriter,r *http.Request){id,ok:=platformID(w,r);if !ok{return};var req tenantStatusRequest;if !httpx.DecodeJSON(w,r,&req){return};s:=tenants.Status(strings.ToLower(strings.TrimSpace(req.Status)));if s!=tenants.StatusActive&&s!=tenants.StatusSuspended{httpx.WriteError(w,httpx.ErrValidation.WithMessage("status must be active or suspended"));return};if err:=h.repo.UpdateStatus(r.Context(),id,s);err==tenants.ErrNotFound{httpx.WriteError(w,httpx.ErrNotFound);return}else if err!=nil{httpx.WriteError(w,httpx.ErrInternal);return};t,_:=h.repo.ByID(r.Context(),id);httpx.WriteJSON(w,http.StatusOK,tenantDTO(t))}
func platformID(w http.ResponseWriter,r *http.Request)(int64,bool){id,err:=strconv.ParseInt(r.PathValue("id"),10,64);if err!=nil||id<1{httpx.WriteError(w,httpx.ErrValidation.WithMessage("invalid tenant id"));return 0,false};return id,true}
func validSlug(s string)bool{if len(s)<3||len(s)>50{return false};for _,r:=range s{if !(unicode.IsLower(r)||unicode.IsDigit(r)||r=='-'){return false}};return true}
