-- Local approval integrity indexes. Database-per-tenant isolation removes tenant_id.
CREATE INDEX idx_workflow_steps_workflow_order
    ON approval_workflow_steps (workflow_id, step_order);
CREATE INDEX idx_gatepass_approvals_status
    ON gatepass_approvals (status, step_order);
CREATE INDEX idx_memberships_user_role_status
    ON tenant_memberships (user_id, role_id, status);
