package visitors

import (
	"context"
	"errors"

	"gatepass/internal/audit"
	"gatepass/internal/httpx"
	"gatepass/internal/settings"
)

var (
	ErrPreRegistrationDisabled=errors.New("visitors: pre-registration is not enabled for this tenant")
	ErrIDNumberRequired=errors.New("visitors: id_number is required for this id type")
	ErrInvalidIDType=errors.New("visitors: id type not found or inactive")
	ErrInvalidCompany=errors.New("visitors: company not found or inactive")
)

type Service struct{
	repo *Repository
	idTypes *IDTypeRepository
	companies *CompanyRepository
	settingsRepo *settings.Repository
	auditRepo *audit.Repository
}

func NewService(repo *Repository,idTypes *IDTypeRepository,companies *CompanyRepository,settingsRepo *settings.Repository,auditRepo *audit.Repository)*Service{return &Service{repo:repo,idTypes:idTypes,companies:companies,settingsRepo:settingsRepo,auditRepo:auditRepo}}

func (s *Service) Create(ctx context.Context,tenantID int64,in CreateInput,actorUserID int64)(*Visitor,error){
	idType,err:=s.idTypes.ByID(ctx,in.IDTypeID);if err!=nil||!idType.Active{return nil,ErrInvalidIDType}
	if idType.RequiresNumber&&(in.IDNumber==nil||*in.IDNumber==""){return nil,ErrIDNumberRequired}
	if in.CompanyID!=nil{c,err:=s.companies.ByID(ctx,*in.CompanyID);if err!=nil||!c.Active{return nil,ErrInvalidCompany}}
	source:=SourceWalkIn
	if in.WantsPreRegistration{if !s.settingsRepo.GetBool(ctx,settings.KeyVisitorsAllowPreRegistration,false){return nil,ErrPreRegistrationDisabled};source=SourcePreRegistered}
	v:=&Visitor{FirstName:in.FirstName,LastName:in.LastName,IDTypeID:in.IDTypeID,IDNumber:in.IDNumber,CompanyID:in.CompanyID,Phone:in.Phone,Email:in.Email,Notes:in.Notes,PhotoRef:in.PhotoRef,Source:source,CreatedBy:&actorUserID}
	id,err:=s.repo.Create(ctx,v);if err!=nil{return nil,err};v.ID=id
	_ = s.auditRepo.Record(ctx,s.auditRepo.DB(),audit.Entry{ActorUserID:&actorUserID,Action:audit.ActionVisitorCreated,EntityType:"visitor",EntityID:&id,Metadata:map[string]any{"source":string(source)}})
	return v,nil
}

func (s *Service) Update(ctx context.Context,tenantID,id int64,in UpdateInput,actorUserID int64)(*Visitor,error){
	existing,err:=s.repo.ByID(ctx,id);if err!=nil{return nil,err}
	effectiveIDTypeID:=existing.IDTypeID;if in.IDTypeID!=nil{effectiveIDTypeID=*in.IDTypeID}
	idType,err:=s.idTypes.ByID(ctx,effectiveIDTypeID);if err!=nil||!idType.Active{return nil,ErrInvalidIDType}
	effectiveIDNumber:=existing.IDNumber;if in.IDNumber!=nil{effectiveIDNumber=in.IDNumber}
	if idType.RequiresNumber&&(effectiveIDNumber==nil||*effectiveIDNumber==""){return nil,ErrIDNumberRequired}
	if in.CompanyID!=nil{c,err:=s.companies.ByID(ctx,*in.CompanyID);if err!=nil||!c.Active{return nil,ErrInvalidCompany}}
	v,err:=s.repo.Update(ctx,id,in,actorUserID);if err!=nil{return nil,err}
	_ = s.auditRepo.Record(ctx,s.auditRepo.DB(),audit.Entry{ActorUserID:&actorUserID,Action:audit.ActionVisitorUpdated,EntityType:"visitor",EntityID:&id})
	return v,nil
}

func (s *Service) Get(ctx context.Context,tenantID,id int64)(*Visitor,error){return s.repo.ByID(ctx,id)}
func (s *Service) List(ctx context.Context,tenantID int64,f ListFilter,p httpx.Pagination)([]Visitor,int,error){return s.repo.List(ctx,f,p)}
func (s *Service) SetBlacklist(ctx context.Context,tenantID,id int64,in BlacklistInput,actorUserID int64)error{
	if err:=s.repo.SetBlacklist(ctx,id,in.Blacklisted,in.Reason,actorUserID);err!=nil{return err}
	_ = s.auditRepo.Record(ctx,s.auditRepo.DB(),audit.Entry{ActorUserID:&actorUserID,Action:"VISITOR_BLACKLIST_UPDATED",EntityType:"visitor",EntityID:&id,Metadata:map[string]any{"blacklisted":in.Blacklisted,"reason":in.Reason}})
	return nil
}
