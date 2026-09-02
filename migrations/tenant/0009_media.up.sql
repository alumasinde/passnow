-- Phase 5: tenant media library.
-- Each tenant owns its own database, so no tenant_id is stored here.

CREATE TABLE IF NOT EXISTS media_files (
    id BIGINT NOT NULL AUTO_INCREMENT,
    public_id CHAR(32) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    storage_path VARCHAR(500) NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    size_bytes BIGINT UNSIGNED NOT NULL,
    purpose VARCHAR(50) NOT NULL DEFAULT 'general',
    created_by BIGINT NOT NULL,
    created_at DATETIME NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_media_files_public_id (public_id),
    KEY idx_media_files_created_at (created_at),
    KEY idx_media_files_purpose (purpose)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT IGNORE INTO permissions (code, label) VALUES
    ('settings.media', 'Manage tenant media files');

INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.code = 'settings.media'
WHERE r.name = 'Tenant Admin';