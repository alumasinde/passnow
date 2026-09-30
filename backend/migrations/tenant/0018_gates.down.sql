DROP TABLE IF EXISTS gates;

DELETE rp FROM role_permissions rp
JOIN permissions p ON p.id=rp.permission_id
WHERE p.code IN ('gate.read','gate.create','gate.update');

DELETE FROM permissions WHERE code IN ('gate.read','gate.create','gate.update');