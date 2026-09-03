package gates

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"gatepass/internal/httpx"
	"gatepass/internal/reqctx"
)

type Handler struct{ repo *Repository }
func NewHandler(repo *Repository)*Handler{return &Handler{repo:repo}}

func (h *Handler) List(w http.ResponseWriter,r *http.Request){
	if _,ok:=reqctx.TenantFromContext(r.Context());!ok{httpx.WriteError(w,httpx.ErrAuthRequired);return}
	items,err:=h.repo.List(r.Context(),r.URL.Query().Get("all")=="true");if err!=nil{httpx.WriteError(w,httpx.ErrInternal);return};httpx.WriteJSON(w,http.StatusOK,items)
}
func (h *Handler) Get(w http.ResponseWriter,r *http.Request){
	if _,ok:=reqctx.TenantFromContext(r.Context());!ok{httpx.WriteError(w,httpx.ErrAuthRequired);return};id,err:=strconv.ParseInt(r.PathValue("id"),10,64);if err!=nil||id<1{httpx.WriteError(w,httpx.ErrNotFound);return};g,err:=h.repo.ByID(r.Context(),id);if errors.Is(err,ErrNotFound){httpx.WriteError(w,httpx.ErrNotFound);return};if err!=nil{httpx.WriteError(w,httpx.ErrInternal);return};httpx.WriteJSON(w,http.StatusOK,g)
}
func validate(in Input)error{
	in.Code=strings.ToUpper(strings.TrimSpace(in.Code));if in.Code==""{return errors.New("gate code is required")};if len(in.Code)>30{return errors.New("gate code is too long")}
	if strings.TrimSpace(in.Name)==""{return errors.New("gate name is required")};if len(strings.TrimSpace(in.Name))>120{return errors.New("gate name is too long")}
	if in.AllowsEntry!=nil&&in.AllowsExit!=nil&&!*in.AllowsEntry&&!*in.AllowsExit{return errors.New("a gate must allow entry, exit, or both")}
	return nil
}
func (h *Handler) Create(w http.ResponseWriter,r *http.Request){if _,ok:=reqctx.TenantFromContext(r.Context());!ok{httpx.WriteError(w,httpx.ErrAuthRequired);return};var in Input;if !httpx.DecodeJSON(w,r,&in){return};if err:=validate(in);err!=nil{httpx.WriteError(w,httpx.ErrValidation.WithMessage(err.Error()));return};g,err:=h.repo.Create(r.Context(),in);if err!=nil{httpx.WriteError(w,httpx.ErrValidation.WithMessage(err.Error()));return};httpx.WriteJSON(w,http.StatusCreated,g)}
func (h *Handler) Update(w http.ResponseWriter,r *http.Request){if _,ok:=reqctx.TenantFromContext(r.Context());!ok{httpx.WriteError(w,httpx.ErrAuthRequired);return};id,err:=strconv.ParseInt(r.PathValue("id"),10,64);if err!=nil||id<1{httpx.WriteError(w,httpx.ErrNotFound);return};var in Input;if !httpx.DecodeJSON(w,r,&in){return};if in.Code!=""||in.Name!=""{if err:=validate(in);err!=nil{httpx.WriteError(w,httpx.ErrValidation.WithMessage(err.Error()));return}};g,err:=h.repo.Update(r.Context(),id,in);if errors.Is(err,ErrNotFound){httpx.WriteError(w,httpx.ErrNotFound);return};if err!=nil{httpx.WriteError(w,httpx.ErrValidation.WithMessage(err.Error()));return};httpx.WriteJSON(w,http.StatusOK,g)}
