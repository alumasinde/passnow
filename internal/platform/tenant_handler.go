package platform

import (
 "context"
 "net/http"
 "strconv"
 "strings"
 "unicode"
 "gatepass/internal/httpx"
 "gatepass/internal/tenants"
)

type TenantHandler struct{ repo *tenants.Repository }
func NewTenantHandler(repo *tenants.Repository)*TenantHandler{return &TenantHandler{repo:repo}}

func domainDTO(d tenants.Domain) map[string]any { return map[string]any{"id":d.ID,"domain":d.Domain,"type":d.Type,"is_primary":d.IsPrimary,"is_verified":d.IsVerified,"created_at":d.CreatedAt} }
func (h *TenantHandler) dto(ctx context.Context,t *tenants.Tenant) map[string]any { ds,_:=h.repo.Domains(ctx,t.ID); domains:=make([]map[string]any,0,len(ds));for _,d:=range ds{domains=append(domains,domainDTO(d))};return map[string]any{"id":t.ID,"name":t.Name,"slug":t.Slug,"status":t.Status,"custom_domain":t.CustomDomain,"custom_domain_verified":t.CustomDomainVerified,"domains":domains,"created_at":t.CreatedAt,"updated_at":t.UpdatedAt} }

func (h *TenantHandler) List(w http.ResponseWriter,r *http.Request){items,err:=h.repo.List(r.Context());if err!=nil{httpx.WriteError(w,httpx.ErrInternal);return};out:=make([]map[string]any,0,len(items));for _,t:=range items{out=append(out,h.dto(r.Context(),t))};httpx.WriteJSON(w,http.StatusOK,map[string]any{"tenants":out})}
func (h *TenantHandler) Get(w http.ResponseWriter,r *http.Request){id,ok:=platformID(w,r);if !ok{return};t,err:=h.repo.ByID(r.Context(),id);if err==tenants.ErrNotFound{httpx.WriteError(w,httpx.ErrNotFound);return};if err!=nil{httpx.WriteError(w,httpx.ErrInternal);return};httpx.WriteJSON(w,http.StatusOK,h.dto(r.Context(),t))}

type tenantStatusRequest struct{Status string `json:"status"`}
func (h *TenantHandler) UpdateStatus(w http.ResponseWriter,r *http.Request){id,ok:=platformID(w,r);if !ok{return};var req tenantStatusRequest;if !httpx.DecodeJSON(w,r,&req){return};s:=tenants.Status(strings.ToLower(strings.TrimSpace(req.Status)));if s!=tenants.StatusActive&&s!=tenants.StatusSuspended{httpx.WriteError(w,httpx.ErrValidation.WithMessage("status must be active or suspended"));return};if err:=h.repo.UpdateStatus(r.Context(),id,s);err==tenants.ErrNotFound{httpx.WriteError(w,httpx.ErrNotFound);return}else if err!=nil{httpx.WriteError(w,httpx.ErrInternal);return};t,_:=h.repo.ByID(r.Context(),id);httpx.WriteJSON(w,http.StatusOK,h.dto(r.Context(),t))}

type tenantUpdateRequest struct{Name string `json:"name"`;Slug string `json:"slug"`}
func (h *TenantHandler) Update(w http.ResponseWriter,r *http.Request){id,ok:=platformID(w,r);if !ok{return};var req tenantUpdateRequest;if !httpx.DecodeJSON(w,r,&req){return};req.Name=strings.TrimSpace(req.Name);req.Slug=strings.ToLower(strings.TrimSpace(req.Slug));if req.Name==""||!validSlug(req.Slug){httpx.WriteError(w,httpx.ErrValidation.WithMessage("organization name and valid slug are required"));return};if err:=h.repo.Update(r.Context(),id,req.Name,req.Slug);err!=nil{httpx.WriteError(w,httpx.ErrInternal);return};t,_:=h.repo.ByID(r.Context(),id);httpx.WriteJSON(w,http.StatusOK,h.dto(r.Context(),t))}

type domainRequest struct{Domain string `json:"domain"`}
func (h *TenantHandler) AddCustomDomain(w http.ResponseWriter,r *http.Request){id,ok:=platformID(w,r);if !ok{return};var req domainRequest;if !httpx.DecodeJSON(w,r,&req){return};d:=strings.ToLower(strings.TrimSpace(req.Domain));if d==""||strings.Contains(d,"/"){httpx.WriteError(w,httpx.ErrValidation.WithMessage("a valid domain is required"));return};if err:=h.repo.AddDomain(r.Context(),id,d,tenants.DomainCustom,false,false);err!=nil{httpx.WriteError(w,httpx.ErrValidation.WithMessage("domain already exists or could not be added"));return};t,_:=h.repo.ByID(r.Context(),id);httpx.WriteJSON(w,http.StatusCreated,h.dto(r.Context(),t))}
func (h *TenantHandler) SetPrimaryDomain(w http.ResponseWriter,r *http.Request){id,ok:=platformID(w,r);if !ok{return};did,err:=strconv.ParseInt(r.PathValue("domainID"),10,64);if err!=nil||did<1{httpx.WriteError(w,httpx.ErrValidation.WithMessage("invalid domain"));return};if err:=h.repo.SetPrimaryDomain(r.Context(),id,did);err!=nil{httpx.WriteError(w,httpx.ErrNotFound);return};t,_:=h.repo.ByID(r.Context(),id);httpx.WriteJSON(w,http.StatusOK,h.dto(r.Context(),t))}
func (h *TenantHandler) DeleteDomain(w http.ResponseWriter,r *http.Request){id,ok:=platformID(w,r);if !ok{return};did,err:=strconv.ParseInt(r.PathValue("domainID"),10,64);if err!=nil||did<1{httpx.WriteError(w,httpx.ErrValidation.WithMessage("invalid domain"));return};if err:=h.repo.DeleteDomain(r.Context(),id,did);err!=nil{httpx.WriteError(w,httpx.ErrInternal);return};w.WriteHeader(http.StatusNoContent)}

func platformID(w http.ResponseWriter,r *http.Request)(int64,bool){id,err:=strconv.ParseInt(r.PathValue("id"),10,64);if err!=nil||id<1{httpx.WriteError(w,httpx.ErrValidation.WithMessage("invalid tenant id"));return 0,false};return id,true}
func validSlug(s string)bool{if len(s)<3||len(s)>50{return false};for _,r:=range s{if !(unicode.IsLower(r)||unicode.IsDigit(r)||r=='-'){return false}};return true}
