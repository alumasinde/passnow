-- RBAC Phase 5 rollback: remove template roles. Custom tenant roles are untouched.
DELETE rp FROM role_permissions rp
JOIN roles r ON r.id=rp.role_id
WHERE r.name IN ('Owner','General Manager','HR Manager','Department Head','Security Manager','Security Supervisor','Gate Officer','Receptionist','Employee','Auditor');
DELETE FROM roles
WHERE name IN ('Owner','General Manager','HR Manager','Department Head','Security Manager','Security Supervisor','Gate Officer','Receptionist','Employee','Auditor')
AND is_system=1;