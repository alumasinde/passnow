-- RBAC Phase 5: organization role templates.
-- Templates are seeded as editable tenant roles. Permission changes remain tenant-owned.
-- Names intentionally describe organizational responsibility rather than implementation details.

INSERT INTO roles (name, is_system) VALUES
('Owner',1),('General Manager',1),('HR Manager',1),('Department Head',1),
('Security Manager',1),('Security Supervisor',1),('Gate Officer',1),
('Receptionist',1),('Employee',1),('Auditor',1)
ON DUPLICATE KEY UPDATE is_system=VALUES(is_system);

-- Start from a clean template mapping so rerunning migrations is deterministic.
DELETE rp FROM role_permissions rp
JOIN roles r ON r.id=rp.role_id
WHERE r.name IN ('Owner','General Manager','HR Manager','Department Head','Security Manager','Security Supervisor','Gate Officer','Receptionist','Employee','Auditor');

-- Owner and General Manager: full organization administration.
INSERT IGNORE INTO role_permissions (role_id,permission_id)
SELECT r.id,p.id FROM roles r JOIN permissions p
WHERE r.name IN ('Owner','General Manager');

-- HR Manager: people, memberships and department administration, plus organization visibility.
INSERT IGNORE INTO role_permissions (role_id,permission_id)
SELECT r.id,p.id FROM roles r JOIN permissions p
WHERE r.name='HR Manager' AND (
p.code LIKE 'employee.%' OR p.code IN ('user.read.all','user.create','user.update.all','user.deactivate',
'membership.read','membership.assign_role','membership.assign_department','membership.update',
'role.read','permission.read','department.read','department.create','department.update',
'visitor.read.all','visit.read.all','gatepass.read.all','report.read.all','report.export.all','audit_log.read.all'));

-- Department Head: department-scoped operations and approvals.
INSERT IGNORE INTO role_permissions (role_id,permission_id)
SELECT r.id,p.id FROM roles r JOIN permissions p
WHERE r.name='Department Head' AND (
p.code LIKE 'visitor.read.department' OR p.code LIKE 'visitor.update.department' OR p.code='visitor.create' OR
p.code LIKE 'visit.read.department' OR p.code LIKE 'visit.update.department' OR p.code='visit.create' OR p.code='visit.cancel.department' OR
p.code LIKE 'gatepass.read.department' OR p.code LIKE 'gatepass.update.department' OR p.code='gatepass.create' OR p.code='gatepass.cancel.department' OR
p.code IN ('approval.read.assigned','approval.read.department','approval.approve','approval.reject',
'employee.read.department','employee.update.department','department.read','report.read.department','report.export.department','audit_log.read.department'));

-- Security management roles.
INSERT IGNORE INTO role_permissions (role_id,permission_id)
SELECT r.id,p.id FROM roles r JOIN permissions p
WHERE r.name='Security Manager' AND (
p.code LIKE 'visitor.%' OR p.code LIKE 'visit.%' OR p.code LIKE 'gatepass.%' OR
p.code IN ('approval.read.department','approval.read.all','approval.approve','approval.reject',
'employee.read.all','department.read','report.read.all','report.export.all','audit_log.read.all'));

INSERT IGNORE INTO role_permissions (role_id,permission_id)
SELECT r.id,p.id FROM roles r JOIN permissions p
WHERE r.name='Security Supervisor' AND (
p.code IN ('visitor.read.all','visitor.create','visitor.update.all','visitor.blacklist',
'visit.read.all','visit.create','visit.update.all','visit.cancel.all','visit.check_in','visit.check_out',
'gatepass.read.all','gatepass.create','gatepass.update.all','gatepass.cancel.all','gatepass.verify','gatepass.check_out','gatepass.check_in',
'approval.read.department','department.read','report.read.department','audit_log.read.department'));

INSERT IGNORE INTO role_permissions (role_id,permission_id)
SELECT r.id,p.id FROM roles r JOIN permissions p
WHERE r.name='Gate Officer' AND p.code IN (
'visitor.read.all','visitor.create','visitor.update.all',
'visit.read.all','visit.create','visit.update.all','visit.check_in','visit.check_out',
'gatepass.read.all','gatepass.verify','gatepass.check_out','gatepass.check_in','department.read');

-- Reception: visitor registration and visit lifecycle, without security administration.
INSERT IGNORE INTO role_permissions (role_id,permission_id)
SELECT r.id,p.id FROM roles r JOIN permissions p
WHERE r.name='Receptionist' AND p.code IN (
'visitor.read.all','visitor.create','visitor.update.all',
'visit.read.all','visit.create','visit.update.all','visit.cancel.all','visit.check_in','visit.check_out','department.read');

-- Employee: self-service and own operational records.
INSERT IGNORE INTO role_permissions (role_id,permission_id)
SELECT r.id,p.id FROM roles r JOIN permissions p
WHERE r.name='Employee' AND p.code IN (
'user.read.own','user.update.own','visitor.read.own','visitor.create','visitor.update.own',
'visit.read.own','visit.create','visit.update.own','visit.cancel.own',
'gatepass.read.own','gatepass.create','gatepass.update.own','gatepass.cancel.own',
'approval.read.assigned','report.read.own');

-- Auditor: read-only organization oversight.
INSERT IGNORE INTO role_permissions (role_id,permission_id)
SELECT r.id,p.id FROM roles r JOIN permissions p
WHERE r.name='Auditor' AND (
p.code IN ('visitor.read.all','visit.read.all','gatepass.read.all','approval.read.all',
'employee.read.all','user.read.all','membership.read','role.read','permission.read',
'department.read','workflow.read','report.read.all','audit_log.read.all'));