CREATE TABLE gates (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(30) NOT NULL,
    name VARCHAR(120) NOT NULL,
    description VARCHAR(255) NULL,
    location VARCHAR(120) NULL,
    allows_entry TINYINT(1) NOT NULL DEFAULT 1,
    allows_exit TINYINT(1) NOT NULL DEFAULT 1,
    is_default TINYINT(1) NOT NULL DEFAULT 0,
    active TINYINT(1) NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL,
    UNIQUE KEY uq_gates_code (code),
    KEY idx_gates_active (active),
    KEY idx_gates_default (is_default)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT IGNORE INTO permissions (code,label) VALUES
('gate.read','View gates'),
('gate.create','Create gates'),
('gate.update','Update gates');

INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id,p.id FROM roles r JOIN permissions p
WHERE r.name='Tenant Admin' AND r.is_system=1 AND p.code IN ('gate.read','gate.create','gate.update');

INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id,p.id FROM roles r JOIN permissions p
WHERE r.name IN ('Owner','General Manager','Security Manager','Security Supervisor') AND p.code IN ('gate.read','gate.create','gate.update');