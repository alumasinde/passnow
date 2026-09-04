package approvals

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gatepass/internal/testutil"
)

func TestWorkflowCreateReadUpdateLifecycle(t *testing.T) {
	db := testutil.OpenMySQL(t)
	for _, table := range []string{"approval_workflows", "approval_workflow_steps", "roles"} { testutil.RequireTable(t, db, table) }
	ctx := context.Background()
	roleID := testutil.MustQueryInt(t, db, "SELECT id FROM roles LIMIT 1")
	if roleID < 1 { t.Skip("no role available in configured MySQL test database") }
	required := true
	repo := NewRepository(db)
	name := fmt.Sprintf("Sprint 3 workflow %d", time.Now().UnixNano())
	id, err := repo.CreateWithSteps(ctx, 1, CreateWorkflowInput{Name:name, Steps:[]StepInput{
		{Label:"First approval", ApproverType:string(ApproverRole), RoleID:&roleID, Required:&required},
		{Label:"Final approval", ApproverType:string(ApproverRole), RoleID:&roleID, Required:&required},
	}})
	if err != nil { t.Fatalf("CreateWithSteps: %v", err) }
	t.Cleanup(func(){ _, _ = db.ExecContext(ctx, "DELETE FROM approval_workflows WHERE id = ?", id) })
	w, steps, err := repo.ByID(ctx, 1, id)
	if err != nil { t.Fatalf("ByID: %v", err) }
	if w.Name != name || len(steps) != 2 || steps[0].StepOrder != 1 || steps[1].StepOrder != 2 { t.Fatalf("workflow mismatch: workflow=%+v steps=%+v", w, steps) }
	updated := name + " updated"
	if err := repo.UpdateWithSteps(ctx, 1, id, UpdateWorkflowInput{Name:updated, Steps:[]StepInput{{Label:"Single final", ApproverType:string(ApproverRole), RoleID:&roleID, Required:&required}}}); err != nil { t.Fatalf("UpdateWithSteps: %v", err) }
	w, steps, err = repo.ByID(ctx, 1, id)
	if err != nil { t.Fatalf("ByID after update: %v", err) }
	if w.Name != updated || len(steps) != 1 || steps[0].Label != "Single final" { t.Fatalf("updated workflow mismatch: workflow=%+v steps=%+v", w, steps) }
}
