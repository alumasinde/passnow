-- Batch 2: Approval integrity.
-- The application validates tenant ownership before inserting workflow steps.
-- These indexes support the tenant-safe lookups and approver work queue.

CREATE INDEX idx_workflow_steps_tenant_workflow_order
    ON approval_workflow_steps (tenant_id, workflow_id, step_order);

CREATE INDEX idx_gatepass_approvals_tenant_status
    ON gatepass_approvals (tenant_id, status, step_order);

CREATE INDEX idx_memberships_tenant_user_role_status
    ON tenant_memberships (tenant_id, user_id, role_id, status);
