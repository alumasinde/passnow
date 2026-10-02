INSERT IGNORE INTO permissions (code,label) VALUES
('media.read','View the tenant media library'),
('media.create','Upload media files'),
('media.delete','Delete media files'),
('settings.update','Change tenant branding and theme');

INSERT IGNORE INTO role_permissions (role_id,permission_id)
SELECT r.id,p.id FROM roles r JOIN permissions p
WHERE r.name IN ('Tenant Admin','Owner','General Manager') AND r.is_system=1
  AND p.code IN ('media.read','media.create','media.delete','settings.update');