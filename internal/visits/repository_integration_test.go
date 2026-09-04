package visits

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gatepass/internal/testutil"
)

func TestVisitRepositoryCheckInCheckOutAndMovements(t *testing.T) {
	db := testutil.OpenMySQL(t)
	for _, table := range []string{"visitors","id_types","visits","gates","users"} { testutil.RequireTable(t, db, table) }
	var movementTable string
	hasMovements := db.QueryRowContext(ctx, "SELECT TABLE_NAME FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'visit_movements'").Scan(&movementTable) == nil
	ctx := context.Background()
	idTypeID := testutil.MustQueryInt(t, db, "SELECT id FROM id_types WHERE active=1 AND deleted_at IS NULL LIMIT 1")
	gateID := testutil.MustQueryInt(t, db, "SELECT id FROM gates WHERE active=1 AND deleted_at IS NULL LIMIT 1")
	userID := testutil.MustQueryInt(t, db, "SELECT id FROM users WHERE deleted_at IS NULL LIMIT 1")
	if idTypeID < 1 || gateID < 1 || userID < 1 { t.Skip("required fixture data unavailable") }
	uniq := time.Now().UnixNano()
	res := testutil.MustExec(t, db, "INSERT INTO visitors(first_name,last_name,id_type_id,id_number,source,status,created_by,created_at,updated_at) VALUES(?,?,?,?, 'walk_in','active',?,NOW(),NOW())", "Sprint","Visitor",idTypeID,fmt.Sprintf("S3-%d",uniq),userID)
	visitorID,_ := res.LastInsertId()
	t.Cleanup(func(){ _, _ = db.ExecContext(ctx, "DELETE FROM visitors WHERE id=?", visitorID) })
	repo := NewRepository(db)
	purpose := "Sprint 3 lifecycle"
	visitID, err := repo.Create(ctx,&Visit{VisitorID:visitorID,EntrySource:EntrySourcePreRegistered,Purpose:&purpose,Status:StatusScheduled,CreatedBy:&userID})
	if err != nil { t.Fatalf("Create: %v",err) }
	t.Cleanup(func(){ _, _ = db.ExecContext(ctx,"DELETE FROM visits WHERE id=?",visitID) })
	checkedIn,err:=repo.CheckIn(ctx,visitID,userID); if err!=nil{t.Fatalf("CheckIn: %v",err)}
	if checkedIn.Status!=StatusCheckedIn || checkedIn.BadgeNumber==nil || checkedIn.BadgeToken==nil || checkedIn.CheckedInAt==nil { t.Fatalf("check-in mismatch: %+v",checkedIn) }
	if hasMovements { if err:=repo.RecordMovement(ctx,visitID,MovementCheckIn,userID,MovementInput{GateID:gateID});err!=nil{t.Fatalf("RecordMovement check-in: %v",err)} }
	if _,err:=repo.CheckIn(ctx,visitID,userID);err!=ErrInvalidTransition{t.Fatalf("second check-in error=%v",err)}
	checkedOut,err:=repo.CheckOut(ctx,visitID,userID);if err!=nil{t.Fatalf("CheckOut: %v",err)}
	if checkedOut.Status!=StatusCheckedOut || checkedOut.CheckedOutAt==nil{t.Fatalf("check-out mismatch: %+v",checkedOut)}
	if hasMovements {
		if err:=repo.RecordMovement(ctx,visitID,MovementCheckOut,userID,MovementInput{GateID:gateID});err!=nil{t.Fatalf("RecordMovement check-out: %v",err)}
		movements,err:=repo.Movements(ctx,visitID);if err!=nil{t.Fatalf("Movements: %v",err)}
		if len(movements)!=2{t.Fatalf("movement count=%d, want 2",len(movements))}
	} else { t.Log("visit_movements table is not present in this database; visit lifecycle tested, movement persistence requires the pending schema migration") }
}
