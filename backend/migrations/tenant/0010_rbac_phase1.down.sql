-- RBAC Phase 1 down migration.
-- Only canonical Phase 1 permissions are removed. Existing legacy permissions remain.
DELETE rp FROM role_permissions rp
JOIN permissions p ON p.id=rp.permission_id
WHERE p.code LIKE 'visitor.%' OR p.code LIKE 'visit.%' OR p.code LIKE 'gatepass.%'
   OR p.code LIKE 'approval.%' OR p.code LIKE 'employee.%' OR p.code LIKE 'user.%'
   OR p.code LIKE 'membership.%' OR p.code LIKE 'role.%' OR p.code LIKE 'permission.%'
   OR p.code LIKE 'department.%' OR p.code LIKE 'workflow.%' OR p.code LIKE 'report.%'
   OR p.code LIKE 'audit_log.%';
DELETE FROM permissions
WHERE code LIKE 'visitor.%' OR code LIKE 'visit.%' OR code LIKE 'gatepass.%'
   OR code LIKE 'approval.%' OR code LIKE 'employee.%' OR code LIKE 'user.%'
   OR code LIKE 'membership.%' OR code LIKE 'role.%' OR code LIKE 'permission.%'
   OR code LIKE 'department.%' OR code LIKE 'workflow.%' OR code LIKE 'report.%'
   OR code LIKE 'audit_log.%';