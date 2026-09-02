-- RBAC Phase 1: canonical permission catalog.
-- Canonical grammar: resource.action.scope, with scope omitted when not meaningful.
-- Legacy permission codes are intentionally retained for backward compatibility;
-- later phases migrate route enforcement and scoped data access.

INSERT IGNORE INTO permissions (code,label) VALUES
('visitor.read.own','View visitors created by the current user'),
('visitor.read.department','View visitors in the current user department'),
('visitor.read.all','View all visitors'),
('visitor.create','Register visitors'),
('visitor.update.own','Update visitors created by the current user'),
('visitor.update.department','Update visitors in the current user department'),
('visitor.update.all','Update any visitor'),
('visitor.blacklist','Blacklist or restore visitors'),

('visit.read.own','View visits created by the current user'),
('visit.read.department','View visits for the current user department'),
('visit.read.all','View all visits'),
('visit.create','Create visits'),
('visit.update.own','Update visits created by the current user'),
('visit.update.department','Update visits for the current user department'),
('visit.update.all','Update any visit'),
('visit.cancel.own','Cancel visits created by the current user'),
('visit.cancel.department','Cancel visits for the current user department'),
('visit.cancel.all','Cancel any visit'),
('visit.check_in','Check visitors in'),
('visit.check_out','Check visitors out'),

('gatepass.read.own','View gatepasses created by the current user'),
('gatepass.read.department','View gatepasses for the current user department'),
('gatepass.read.all','View all gatepasses'),
('gatepass.create','Create gatepasses'),
('gatepass.update.own','Update gatepasses created by the current user'),
('gatepass.update.department','Update gatepasses for the current user department'),
('gatepass.update.all','Update any gatepass'),
('gatepass.cancel.own','Cancel gatepasses created by the current user'),
('gatepass.cancel.department','Cancel gatepasses for the current user department'),
('gatepass.cancel.all','Cancel any gatepass'),
('gatepass.verify','Verify gatepasses at the gate'),
('gatepass.check_out','Record gatepass outbound movement'),
('gatepass.check_in','Record gatepass return movement'),

('approval.read.assigned','View approvals assigned to the current user'),
('approval.read.department','View approvals for the current user department'),
('approval.read.all','View all approvals'),
('approval.approve','Approve assigned approval steps'),
('approval.reject','Reject assigned approval steps'),
('approval.reassign','Reassign approval work'),

('employee.read.department','View employees in the current user department'),
('employee.read.all','View all employees'),
('employee.create','Create employee records'),
('employee.update.department','Update employees in the current user department'),
('employee.update.all','Update any employee record'),
('employee.deactivate','Deactivate employee records'),

('user.read.own','View own account'),
('user.read.all','View all user accounts'),
('user.create','Create user accounts'),
('user.update.own','Update own account'),
('user.update.all','Update any user account'),
('user.deactivate','Deactivate user accounts'),

('membership.read','View tenant memberships'),
('membership.assign_role','Assign membership roles'),
('membership.assign_department','Assign user departments'),
('membership.update','Update membership status and access'),

('role.read','View roles'),
('role.create','Create roles'),
('role.update','Update roles'),
('role.delete','Delete roles'),
('permission.read','View permission catalog'),
('permission.assign','Assign permissions to roles'),

('department.read','View departments'),
('department.create','Create departments'),
('department.update','Update departments'),
('department.delete','Delete departments'),

('workflow.read','View approval workflows'),
('workflow.create','Create approval workflows'),
('workflow.update','Update approval workflows'),
('workflow.delete','Delete approval workflows'),
('workflow.activate','Activate approval workflows'),
('workflow.deactivate','Deactivate approval workflows'),

('report.read.own','View own reports'),
('report.read.department','View department reports'),
('report.read.all','View organization reports'),
('report.export.department','Export department reports'),
('report.export.all','Export organization reports'),
('audit_log.read.department','View department audit logs'),
('audit_log.read.all','View organization audit logs');

-- Existing Tenant Admins must receive the complete canonical catalog.
INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
WHERE r.name = 'Tenant Admin' AND r.is_system = 1;
