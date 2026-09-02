package visitors

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"

	"gatepass/internal/httpx"
	"gatepass/internal/reqctx"
)

type Handler struct{svc *Service;idTypes *IDTypeRepository;companies *CompanyRepository}
func NewHandler(svc *Service,idTypes *IDTypeRepository,companies *CompanyRepository)*Handler{return &Handler{svc:svc,idTypes:idTypes,companies:companies}}

func tenantRequest(w http.ResponseWriter,r *http.Request)(int64,bool){
	t,ok:=reqctx.TenantFromContext(r.Context());if !ok{httpx.WriteError(w,httpx.ErrAuthRequired);return 0,false};return t.ID,true
}

func (h *Handler) visitorDTO(ctx context.Context,v *Visitor) DTO {
	d:=ToDTO(v)
	if t,err:=h.idTypes.ByID(ctx,v.IDTypeID);err==nil { d.IDTypeName=t.Name }
	if v.CompanyID!=nil {
		if company,err:=h.companies.ByID(ctx,*v.CompanyID);err==nil { d.CompanyName=company.Name }
	}
	return d
}

func visitorStatusOptions() []map[string]string {
	return []map[string]string{
		{"value":string(StatusActive),"label":"Active"},
		{"value":string(StatusBlacklisted),"label":"Blacklisted"},
	}
}

func (h *Handler) Create(w http.ResponseWriter,r *http.Request){
	tenantID,ok:=tenantRequest(w,r);if !ok{return}
	claims,ok:=reqctx.ClaimsFromContext(r.Context());if !ok{httpx.WriteError(w,httpx.ErrAuthRequired);return}
	var in CreateInput;if !httpx.DecodeJSON(w,r,&in){return}
	if in.FirstName==""||in.LastName==""||in.IDTypeID==0{httpx.WriteError(w,httpx.ErrValidation.WithMessage("first_name, last_name and id_type_id are required"));return}
	v,err:=h.svc.Create(r.Context(),tenantID,in,claims.UserID);if err!=nil{writeServiceError(w,err);return}
	httpx.WriteJSON(w,http.StatusCreated,h.visitorDTO(r.Context(),v))
}

func (h *Handler) Get(w http.ResponseWriter,r *http.Request){
	tenantID,ok:=tenantRequest(w,r);if !ok{return};_ = tenantID
	id,err:=parseIDParam(r);if err!=nil{httpx.WriteError(w,httpx.ErrNotFound);return}
	v,err:=h.svc.Get(r.Context(),tenantID,id);if err!=nil{writeServiceError(w,err);return}
	httpx.WriteJSON(w,http.StatusOK,h.visitorDTO(r.Context(),v))
}

func (h *Handler) List(w http.ResponseWriter,r *http.Request){
	tenantID,ok:=tenantRequest(w,r);if !ok{return}
	var f ListFilter;q:=r.URL.Query()
	if s:=q.Get("status");s!=""{st:=Status(s);f.Status=&st}
	if c:=q.Get("company_id");c!=""{if n,err:=strconv.ParseInt(c,10,64);err==nil{f.CompanyID=&n}}
	if b:=q.Get("blacklisted");b!=""{if v,err:=strconv.ParseBool(b);err==nil{f.Blacklisted=&v}}
	f.Search=q.Get("q")
	p:=httpx.ParsePagination(r)
	items,total,err:=h.svc.List(r.Context(),tenantID,f,p);if err!=nil{log.Printf("VISITOR LIST FAILED: %v",err);httpx.WriteError(w,httpx.ErrInternal);return}
	dtos:=make([]DTO,0,len(items));for i:=range items{dtos=append(dtos,h.visitorDTO(r.Context(),&items[i]))}
	httpx.WriteJSON(w,http.StatusOK,map[string]any{
		"items":dtos,"limit":p.Limit,"offset":p.Offset,
		"meta":map[string]any{"total":total,"statuses":visitorStatusOptions()},
	})
}

func (h *Handler) Update(w http.ResponseWriter,r *http.Request){
	tenantID,ok:=tenantRequest(w,r);if !ok{return}
	claims,ok:=reqctx.ClaimsFromContext(r.Context());if !ok{httpx.WriteError(w,httpx.ErrAuthRequired);return}
	id,err:=parseIDParam(r);if err!=nil{httpx.WriteError(w,httpx.ErrNotFound);return}
	var in UpdateInput;if !httpx.DecodeJSON(w,r,&in){return}
	v,err:=h.svc.Update(r.Context(),tenantID,id,in,claims.UserID);if err!=nil{writeServiceError(w,err);return}
	httpx.WriteJSON(w,http.StatusOK,h.visitorDTO(r.Context(),v))
}

func (h *Handler) SetBlacklist(w http.ResponseWriter,r *http.Request){
	tenantID,ok:=tenantRequest(w,r);if !ok{return}
	claims,ok:=reqctx.ClaimsFromContext(r.Context());if !ok{httpx.WriteError(w,httpx.ErrAuthRequired);return}
	id,err:=parseIDParam(r);if err!=nil{httpx.WriteError(w,httpx.ErrNotFound);return}
	var in BlacklistInput;if !httpx.DecodeJSON(w,r,&in){return}
	if in.Blacklisted&&(in.Reason==nil||*in.Reason==""){httpx.WriteError(w,httpx.ErrValidation.WithMessage("reason is required when blacklisting a visitor"));return}
	if err:=h.svc.SetBlacklist(r.Context(),tenantID,id,in,claims.UserID);err!=nil{writeServiceError(w,err);return}
	v,err:=h.svc.Get(r.Context(),tenantID,id);if err!=nil{writeServiceError(w,err);return}
	httpx.WriteJSON(w,http.StatusOK,h.visitorDTO(r.Context(),v))
}

func (h *Handler) ListIDTypes(w http.ResponseWriter,r *http.Request){
	if _,ok:=tenantRequest(w,r);!ok{return};activeOnly:=r.URL.Query().Get("all")!="true"
	items,err:=h.idTypes.List(r.Context(),activeOnly);if err!=nil{httpx.WriteError(w,httpx.ErrInternal);return}
	dtos:=make([]IDTypeDTO,0,len(items));for i:=range items{dtos=append(dtos,IDTypeToDTO(&items[i]))};httpx.WriteJSON(w,http.StatusOK,dtos)
}
func (h *Handler) GetIDType(w http.ResponseWriter,r *http.Request){
	if _,ok:=tenantRequest(w,r);!ok{return};id,err:=parseIDParam(r);if err!=nil{httpx.WriteError(w,httpx.ErrNotFound);return}
	t,err:=h.idTypes.ByID(r.Context(),id);if err!=nil{log.Printf("ID TYPE GET FAILED: id=%d error=%v",id,err);httpx.WriteError(w,httpx.ErrNotFound);return}
	httpx.WriteJSON(w,http.StatusOK,IDTypeToDTO(t))
}
func (h *Handler) CreateIDType(w http.ResponseWriter,r *http.Request){
	if _,ok:=tenantRequest(w,r);!ok{return};var in IDTypeInput;if !httpx.DecodeJSON(w,r,&in){return}
	if in.Name==""||in.Code==""{httpx.WriteError(w,httpx.ErrValidation.WithMessage("name and code are required"));return}
	requires:=true;if in.RequiresNumber!=nil{requires=*in.RequiresNumber}
	active:=true;if in.Active!=nil{active=*in.Active}
	id,err:=h.idTypes.Create(r.Context(),in.Name,in.Code,requires,active);if err!=nil{log.Printf("ID TYPE CREATE FAILED: name=%q code=%q error=%v",in.Name,in.Code,err);writeServiceError(w,err);return}
	t,err:=h.idTypes.ByID(r.Context(),id);if err!=nil{httpx.WriteError(w,httpx.ErrInternal);return};httpx.WriteJSON(w,http.StatusCreated,IDTypeToDTO(t))
}
func (h *Handler) UpdateIDType(w http.ResponseWriter,r *http.Request){
	if _,ok:=tenantRequest(w,r);!ok{return};id,err:=parseIDParam(r);if err!=nil{httpx.WriteError(w,httpx.ErrNotFound);return}
	var in IDTypeInput;if !httpx.DecodeJSON(w,r,&in){return};if err:=h.idTypes.Update(r.Context(),id,in);err!=nil{writeServiceError(w,err);return}
	t,err:=h.idTypes.ByID(r.Context(),id);if err!=nil{writeServiceError(w,err);return};httpx.WriteJSON(w,http.StatusOK,IDTypeToDTO(t))
}

func (h *Handler) ListCompanies(w http.ResponseWriter,r *http.Request){
	if _,ok:=tenantRequest(w,r);!ok{return};activeOnly:=r.URL.Query().Get("all")!="true";p:=httpx.ParsePagination(r)
	items,total,err:=h.companies.List(r.Context(),activeOnly,p);if err!=nil{httpx.WriteError(w,httpx.ErrInternal);return}
	dtos:=make([]CompanyDTO,0,len(items));for i:=range items{dtos=append(dtos,CompanyToDTO(&items[i]))}
	httpx.WriteJSON(w,http.StatusOK,httpx.ListEnvelope[CompanyDTO]{Items:dtos,Limit:p.Limit,Offset:p.Offset,Total:total})
}
func (h *Handler) GetCompany(w http.ResponseWriter,r *http.Request){
	if _,ok:=tenantRequest(w,r);!ok{return}
	id,err:=parseIDParam(r);if err!=nil{httpx.WriteError(w,httpx.ErrNotFound);return}
	company,err:=h.companies.ByID(r.Context(),id);if err!=nil{writeServiceError(w,err);return}
	httpx.WriteJSON(w,http.StatusOK,CompanyToDTO(company))
}
func (h *Handler) CreateCompany(w http.ResponseWriter,r *http.Request){
	if _,ok:=tenantRequest(w,r);!ok{return};var in CompanyInput;if !httpx.DecodeJSON(w,r,&in){return}
	if in.Name==""{httpx.WriteError(w,httpx.ErrValidation.WithMessage("name is required"));return}
	id,err:=h.companies.Create(r.Context(),in);if err!=nil{writeServiceError(w,err);return};c,err:=h.companies.ByID(r.Context(),id);if err!=nil{httpx.WriteError(w,httpx.ErrInternal);return};httpx.WriteJSON(w,http.StatusCreated,CompanyToDTO(c))
}
func (h *Handler) UpdateCompany(w http.ResponseWriter,r *http.Request){
	if _,ok:=tenantRequest(w,r);!ok{return};id,err:=parseIDParam(r);if err!=nil{httpx.WriteError(w,httpx.ErrNotFound);return}
	var in CompanyInput;if !httpx.DecodeJSON(w,r,&in){return};if err:=h.companies.Update(r.Context(),id,in);err!=nil{writeServiceError(w,err);return}
	c,err:=h.companies.ByID(r.Context(),id);if err!=nil{writeServiceError(w,err);return};httpx.WriteJSON(w,http.StatusOK,CompanyToDTO(c))
}

func parseIDParam(r *http.Request)(int64,error){return strconv.ParseInt(r.PathValue("id"),10,64)}
func writeServiceError(w http.ResponseWriter,err error){
	switch{
	case errors.Is(err,ErrNotFound),errors.Is(err,ErrIDTypeNotFound),errors.Is(err,ErrCompanyNotFound):httpx.WriteError(w,httpx.ErrNotFound)
	case errors.Is(err,ErrInvalidIDType):httpx.WriteError(w,httpx.ErrValidation.WithMessage("id_type_id is invalid or inactive"))
	case errors.Is(err,ErrInvalidCompany):httpx.WriteError(w,httpx.ErrValidation.WithMessage("company_id is invalid or inactive"))
	case errors.Is(err,ErrIDNumberRequired):httpx.WriteError(w,httpx.ErrValidation.WithMessage("id_number is required for this id type"))
	case errors.Is(err,ErrDuplicateIDNumber):httpx.WriteError(w,httpx.AppError{Code:"duplicate_id_number",Message:"this ID document is already registered",Status:http.StatusConflict})
	case errors.Is(err,ErrPreRegistrationDisabled):httpx.WriteError(w,httpx.AppError{Code:"pre_registration_disabled",Message:"pre-registration is not enabled for this tenant",Status:http.StatusForbidden})
	case errors.Is(err,ErrCompanyNameTaken):httpx.WriteError(w,httpx.AppError{Code:"company_name_taken",Message:"a company with this name already exists",Status:http.StatusConflict})
	default:httpx.WriteError(w,httpx.ErrInternal)
	}
}
