-- Phase 4: tenant branding and theme permission.
-- Theme values themselves are stored as JSON in tenant_settings so tenants
-- remain isolated by database and the configuration can evolve without
-- repeatedly adding schema columns.

INSERT IGNORE INTO permissions (code, label) VALUES
    ('settings.theme', 'Manage tenant branding and theme');

-- Existing tenants already have a Tenant Admin role. Grant the new setting
-- permission during migration so an administrator does not get locked out
-- after the feature is deployed.
INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.code = 'settings.theme'
WHERE r.name = 'Tenant Admin';