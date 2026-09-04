package gatepasses

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"gatepass/internal/approvals"
	"gatepass/internal/testutil"
)

func TestGatepassApprovalAndReturnableLifecycle(t *testing.T) {
	db := testutil.OpenMySQL(t)
	for _, table := range []string{"gatepasses","gatepass_approvals","gatepass_items","gatepass_types","roles","users","tenant_memberships","number_sequences"} { testutil.RequireTable(t,db,table) }
	ctx:=context.Background()
	actorID:=testutil.MustQueryInt(t,db,"SELECT id FROM users WHERE deleted_at IS NULL LIMIT 1")
	roleID:=testutil.MustQueryInt(t,db,"SELECT role_id FROM tenant_memberships WHERE user_id=? AND status='active' LIMIT 1",actorID)
	if actorID<1||roleID<1{t.Skip("active user membership fixture unavailable")}
	uniq:=time.Now().UnixNano()
	typeRes:=testutil.MustExec(t,db,"INSERT INTO gatepass_types(name,code,direction,is_returnable_default,requires_items,requires_approval,active,created_at,updated_at) VALUES(?,?, 'out',1,0,0,1,NOW(),NOW())","Sprint 3 Type",fmt.Sprintf("S3%d",uniq))
	typeID,_:=typeRes.LastInsertId()
	t.Cleanup(func(){_,_=db.ExecContext(ctx,"DELETE FROM gatepass_types WHERE id=?",typeID)})
	repo:=NewRepository(db,NewItemRepository(db))
	purpose:="Returnable equipment"
	g:=&Gatepass{GatepassTypeID:typeID,RequesterType:RequesterEmployee,RequesterUserID:&actorID,Purpose:&purpose,IsReturnable:true,RequiresApproval:true,CreatedBy:&actorID}
	id,_,err:=repo.Create(ctx,CreateInputResolved{Gatepass:g,Items:[]ItemInput{{Name:"Laptop",Quantity:1,Direction:"leaving"}},WorkflowSteps:[]WorkflowStepSnapshot{{StepOrder:1,Label:"Manager",ApproverType:string(approvals.ApproverRole),RoleID:&roleID,Required:true},{StepOrder:2,Label:"Security",ApproverType:string(approvals.ApproverRole),RoleID:&roleID,Required:true}},NumberScope:"sprint3_gatepass",NumberPrefix:"S3",NumberPeriod:fmt.Sprintf("%d",time.Now().Year())})
	if err!=nil{t.Fatalf("Create: %v",err)}
	t.Cleanup(func(){_,_=db.ExecContext(ctx,"DELETE FROM gatepasses WHERE id=?",id)})
	created,err:=repo.ByID(ctx,id);if err!=nil{t.Fatal(err)}
	if created.Status!=StatusPendingApproval{t.Fatalf("status=%s, want pending_approval",created.Status)}
	steps,err:=repo.ApprovalSteps(ctx,id);if err!=nil||len(steps)!=2{t.Fatalf("ApprovalSteps err=%v len=%d",err,len(steps))}
	if _,err:=repo.ActOnApprovalStep(ctx,id,steps[1].ID,actorID,true,"too early");err==nil{t.Fatal("later step succeeded before earlier required step")}
	if _,err:=repo.ActOnApprovalStep(ctx,id,steps[0].ID,actorID+999999,true,"unauthorized");!errors.Is(err,ErrNotEligibleApprover){t.Fatalf("unauthorized error=%v",err)}
	if _,err:=repo.ActOnApprovalStep(ctx,id,steps[0].ID,actorID,true,"approved");err!=nil{t.Fatalf("approve first: %v",err)}
	mid,err:=repo.ByID(ctx,id);if err!=nil{t.Fatal(err)};if mid.Status!=StatusPendingApproval{t.Fatalf("after first status=%s",mid.Status)}
	approved,err:=repo.ActOnApprovalStep(ctx,id,steps[1].ID,actorID,true,"final");if err!=nil{t.Fatalf("approve final: %v",err)}
	if approved.Status!=StatusApproved{t.Fatalf("status=%s, want approved",approved.Status)}
	out,err:=repo.CheckOut(ctx,id,actorID,string(DirectionOut));if err!=nil{t.Fatalf("CheckOut: %v",err)}
	if out.Status!=StatusAwaitingReturn{t.Fatalf("returnable checkout status=%s, want awaiting_return",out.Status)}
	in,err:=repo.CheckIn(ctx,id,actorID,string(DirectionOut));if err!=nil{t.Fatalf("CheckIn return: %v",err)}
	if in.Status!=StatusCompleted{t.Fatalf("return status=%s, want completed",in.Status)}
}
