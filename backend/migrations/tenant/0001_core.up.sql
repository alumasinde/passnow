-- Tenant database core schema.
-- Each tenant has its own physical database. Tenant isolation is therefore
-- table are intentionally not part of this schema.

CREATE TABLE users (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    -- User accounts are local to this tenant database.
    email           VARCHAR(255) NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    first_name      VARCHAR(100) NOT NULL,
    last_name       VARCHAR(100) NOT NULL,
    status          ENUM('active','disabled') NOT NULL DEFAULT 'active',
    failed_login_count SMALLINT UNSIGNED NOT NULL DEFAULT 0,
    locked_until    DATETIME NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at      DATETIME NULL,

    UNIQUE KEY uq_users_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE roles (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name            VARCHAR(100) NOT NULL,
    is_system       TINYINT(1) NOT NULL DEFAULT 0, -- seeded roles (Tenant Admin, etc.)
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uq_roles_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE permissions (
    id      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    code    VARCHAR(100) NOT NULL, -- e.g. "gatepasses.approve"
    label   VARCHAR(160) NOT NULL,

    UNIQUE KEY uq_permissions_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE role_permissions (
    role_id       BIGINT UNSIGNED NOT NULL,
    permission_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    CONSTRAINT fk_rp_role FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    CONSTRAINT fk_rp_permission FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- A user's local membership and role. The membership is inherently for this
-- tenant because this entire database belongs to one tenant.
CREATE TABLE tenant_memberships (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id     BIGINT UNSIGNED NOT NULL,
    role_id     BIGINT UNSIGNED NOT NULL,
    status      ENUM('active','invited','disabled') NOT NULL DEFAULT 'invited',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uq_membership_user (user_id),
    CONSTRAINT fk_tm_user FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_tm_role FOREIGN KEY (role_id) REFERENCES roles(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE refresh_tokens (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id     BIGINT UNSIGNED NOT NULL,
    token_hash  CHAR(64) NOT NULL, -- SHA-256 of the token; never store raw tokens
    expires_at  DATETIME NOT NULL,
    revoked_at  DATETIME NULL,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY uq_refresh_token_hash (token_hash),
    KEY idx_refresh_user (user_id),
    CONSTRAINT fk_rt_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE audit_logs (
    id           BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    actor_user_id BIGINT UNSIGNED NULL,
    action       VARCHAR(64) NOT NULL,
    entity_type  VARCHAR(64) NOT NULL,
    entity_id    BIGINT UNSIGNED NULL,
    request_id   VARCHAR(64) NULL,
    ip_address   VARCHAR(45) NULL,
    user_agent   VARCHAR(255) NULL,
    metadata     JSON NULL,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    KEY idx_audit_time (created_at),
    KEY idx_audit_entity (entity_type, entity_id)
    -- Audit rows intentionally survive application-level cleanup.
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
