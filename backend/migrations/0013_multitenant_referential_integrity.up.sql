-- Batch 6: production integration / database-level tenant isolation.
-- Repository WHERE tenant_id clauses remain required, but critical tenant-owned
-- relationships are now also enforced by composite foreign keys so a programming
-- mistake cannot connect Tenant A data to Tenant B data.

-- Composite parent keys used by the tenant-bound foreign keys below.
ALTER TABLE roles
    ADD UNIQUE KEY uq_roles_tenant_id (tenant_id, id);

ALTER TABLE approval_workflows
    ADD UNIQUE KEY uq_workflows_tenant_id (tenant_id, id);

ALTER TABLE gatepass_types
    ADD UNIQUE KEY uq_gatepass_types_tenant_id (tenant_id, id);

ALTER TABLE departments
    ADD UNIQUE KEY uq_departments_tenant_id (tenant_id, id);

ALTER TABLE id_types
    ADD UNIQUE KEY uq_id_types_tenant_id (tenant_id, id);

ALTER TABLE visitor_companies
    ADD UNIQUE KEY uq_visitor_companies_tenant_id (tenant_id, id);

ALTER TABLE visitors
    ADD UNIQUE KEY uq_visitors_tenant_id (tenant_id, id);

ALTER TABLE visit_types
    ADD UNIQUE KEY uq_visit_types_tenant_id (tenant_id, id);

ALTER TABLE visits
    ADD UNIQUE KEY uq_visits_tenant_id (tenant_id, id);

ALTER TABLE gatepasses
    ADD UNIQUE KEY uq_gatepasses_tenant_id (tenant_id, id);

ALTER TABLE gatepass_items
    ADD UNIQUE KEY uq_gatepass_items_tenant_id (tenant_id, id);

ALTER TABLE gatepass_movements
    ADD UNIQUE KEY uq_gatepass_movements_tenant_id (tenant_id, id);

-- Membership -> role must be within the same tenant.
ALTER TABLE tenant_memberships
    ADD CONSTRAINT fk_tm_role_same_tenant
    FOREIGN KEY (tenant_id, role_id) REFERENCES roles(tenant_id, id);

-- Workflow step -> role must be within the workflow's tenant.
ALTER TABLE approval_workflow_steps
    ADD CONSTRAINT fk_wfstep_role_same_tenant
    FOREIGN KEY (tenant_id, role_id) REFERENCES roles(tenant_id, id);

-- Workflow step -> specific user remains global-user scoped; membership
-- eligibility is checked at workflow creation and approval time.

-- Gatepass type/workflow must belong to the same tenant.
ALTER TABLE gatepass_types
    ADD CONSTRAINT fk_gptype_workflow_same_tenant
    FOREIGN KEY (tenant_id, workflow_id) REFERENCES approval_workflows(tenant_id, id);

-- Gatepass -> type/department/visitor/visit/workflow must remain in tenant.
ALTER TABLE gatepasses
    ADD CONSTRAINT fk_gp_type_same_tenant
    FOREIGN KEY (tenant_id, gatepass_type_id) REFERENCES gatepass_types(tenant_id, id),
    ADD CONSTRAINT fk_gp_department_same_tenant
    FOREIGN KEY (tenant_id, department_id) REFERENCES departments(tenant_id, id),
    ADD CONSTRAINT fk_gp_visitor_same_tenant
    FOREIGN KEY (tenant_id, requester_visitor_id) REFERENCES visitors(tenant_id, id),
    ADD CONSTRAINT fk_gp_visit_same_tenant
    FOREIGN KEY (tenant_id, visit_id) REFERENCES visits(tenant_id, id);

-- Visitor lookups must stay inside the tenant.
ALTER TABLE visitors
    ADD CONSTRAINT fk_visitor_idtype_same_tenant
    FOREIGN KEY (tenant_id, id_type_id) REFERENCES id_types(tenant_id, id),
    ADD CONSTRAINT fk_visitor_company_same_tenant
    FOREIGN KEY (tenant_id, company_id) REFERENCES visitor_companies(tenant_id, id);

-- Visit lookups must stay inside the tenant.
ALTER TABLE visits
    ADD CONSTRAINT fk_visit_visitor_same_tenant
    FOREIGN KEY (tenant_id, visitor_id) REFERENCES visitors(tenant_id, id),
    ADD CONSTRAINT fk_visit_type_same_tenant
    FOREIGN KEY (tenant_id, visit_type_id) REFERENCES visit_types(tenant_id, id),
    ADD CONSTRAINT fk_visit_department_same_tenant
    FOREIGN KEY (tenant_id, department_id) REFERENCES departments(tenant_id, id);

-- Gatepass items belong to the same tenant as their parent gatepass.
ALTER TABLE gatepass_items
    ADD CONSTRAINT fk_gpitem_gatepass_same_tenant
    FOREIGN KEY (tenant_id, gatepass_id) REFERENCES gatepasses(tenant_id, id);

-- Movement belongs to the same tenant as its gatepass.
ALTER TABLE gatepass_movements
    ADD CONSTRAINT fk_gpm_gatepass_same_tenant
    FOREIGN KEY (tenant_id, gatepass_id) REFERENCES gatepasses(tenant_id, id);

-- Movement item belongs to the same tenant as its gatepass item.
ALTER TABLE gatepass_movement_items
    ADD CONSTRAINT fk_gpmi_item_same_tenant
    FOREIGN KEY (tenant_id, gatepass_item_id) REFERENCES gatepass_items(tenant_id, id);

-- Snapshotted approval belongs to the same tenant as its gatepass.
ALTER TABLE gatepass_approvals
    ADD CONSTRAINT fk_gpapproval_gatepass_same_tenant
    FOREIGN KEY (tenant_id, gatepass_id) REFERENCES gatepasses(tenant_id, id);
