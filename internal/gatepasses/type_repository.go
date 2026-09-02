package gatepasses

import (
    "context"
    "database/sql"
    "errors"
)

var ErrTypeNotFound = errors.New("gatepasses: type not found")

type TypeRepository struct { db *sql.DB }
func NewTypeRepository(db *sql.DB) *TypeRepository { return &TypeRepository{db: db} }

const typeCols = `id, name, code, description, direction, is_returnable_default, returnability_policy, requires_items, requires_approval, workflow_id, active`

func (r *TypeRepository) scan(row interface{ Scan(dest ...any) error }) (*GatepassType, error) {
    var t GatepassType
    if err := row.Scan(&t.ID,&t.Name,&t.Code,&t.Description,&t.Direction,&t.IsReturnableDefault,&t.ReturnabilityPolicy,&t.RequiresItems,&t.RequiresApproval,&t.WorkflowID,&t.Active); err != nil {
        if errors.Is(err,sql.ErrNoRows){return nil,ErrTypeNotFound}; return nil,err
    }
    return &t,nil
}
func (r *TypeRepository) List(ctx context.Context, activeOnly bool) ([]GatepassType,error) {
    q:="SELECT "+typeCols+" FROM gatepass_types WHERE deleted_at IS NULL"; if activeOnly{q+=" AND active = 1"}; q+=" ORDER BY name"
    rows,err:=r.db.QueryContext(ctx,q);if err!=nil{return nil,err};defer rows.Close()
    var out []GatepassType;for rows.Next(){t,err:=r.scan(rows);if err!=nil{return nil,err};out=append(out,*t)};return out,rows.Err()
}
func (r *TypeRepository) ByID(ctx context.Context,id int64)(*GatepassType,error){return r.scan(r.db.QueryRowContext(ctx,"SELECT "+typeCols+" FROM gatepass_types WHERE id = ? AND deleted_at IS NULL LIMIT 1",id))}
func (r *TypeRepository) Create(ctx context.Context,in TypeInput)(int64,error){
    returnable:=false;items:=false;approval:=false;policy:=ReturnabilityOptional
    if in.IsReturnableDefault!=nil{returnable=*in.IsReturnableDefault};if in.ReturnabilityPolicy!=nil{policy=ReturnabilityPolicy(*in.ReturnabilityPolicy)}
    switch policy{case ReturnabilityOptional:case ReturnabilityRequired:returnable=true;case ReturnabilityNotAllowed:returnable=false;default:return 0,errors.New("invalid returnability_policy")}
    if in.RequiresItems!=nil{items=*in.RequiresItems};if in.RequiresApproval!=nil{approval=*in.RequiresApproval}
    if approval && in.WorkflowID==nil{return 0,errors.New("workflow_id is required when requires_approval is true")}
    var workflowID any
    if in.WorkflowID!=nil{workflowID=*in.WorkflowID}else{workflowID=nil}
    res,err:=r.db.ExecContext(ctx,`INSERT INTO gatepass_types (name, code, description, direction, is_returnable_default, returnability_policy, requires_items, requires_approval, workflow_id, active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`,in.Name,in.Code,in.Description,in.Direction,returnable,string(policy),items,approval,workflowID,func() bool { if in.Active != nil { return *in.Active }; return true }());if err!=nil{return 0,err};return res.LastInsertId()
}
func (r *TypeRepository) Update(ctx context.Context,id int64,in TypeInput) error {
    t,err:=r.ByID(ctx,id);if err!=nil{return err}
    if in.Name!=""{t.Name=in.Name};if in.Code!=""{t.Code=in.Code};if in.Description!=nil{t.Description=in.Description};if in.Direction!=""{t.Direction=Direction(in.Direction)}
    if in.IsReturnableDefault!=nil{t.IsReturnableDefault=*in.IsReturnableDefault}
    if in.ReturnabilityPolicy!=nil{policy:=ReturnabilityPolicy(*in.ReturnabilityPolicy);switch policy{case ReturnabilityOptional:case ReturnabilityRequired:t.IsReturnableDefault=true;case ReturnabilityNotAllowed:t.IsReturnableDefault=false;default:return errors.New("invalid returnability_policy")};t.ReturnabilityPolicy=policy}
    if in.RequiresItems!=nil{t.RequiresItems=*in.RequiresItems};if in.RequiresApproval!=nil{t.RequiresApproval=*in.RequiresApproval};if in.Active!=nil{t.Active=*in.Active}
    if !t.RequiresApproval{t.WorkflowID=nil}else if in.WorkflowID!=nil{t.WorkflowID=in.WorkflowID}
    if t.RequiresApproval && t.WorkflowID==nil{return errors.New("workflow_id is required when requires_approval is true")}
    _,err=r.db.ExecContext(ctx,`UPDATE gatepass_types SET name=?, code=?, description=?, direction=?, is_returnable_default=?, returnability_policy=?, requires_items=?, requires_approval=?, workflow_id=?, active=?, updated_at=NOW() WHERE id=?`,t.Name,t.Code,t.Description,t.Direction,t.IsReturnableDefault,t.ReturnabilityPolicy,t.RequiresItems,t.RequiresApproval,t.WorkflowID,t.Active,id);return err
}