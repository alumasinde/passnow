-- Platform administrators are PassNow operators, not tenant users.
-- Roles are intentionally small for Phase 1: owner and admin.
CREATE TABLE IF NOT EXISTS platform_admins (
    user_id     BIGINT UNSIGNED NOT NULL PRIMARY KEY,
    role        ENUM('owner','admin') NOT NULL DEFAULT 'admin',
    status      ENUM('active','disabled') NOT NULL DEFAULT 'active',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_platform_admin_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
