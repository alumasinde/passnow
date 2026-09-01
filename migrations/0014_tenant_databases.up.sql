-- Platform database registry for database-per-tenant architecture.
-- Passwords are encrypted by the application before being stored here.
CREATE TABLE IF NOT EXISTS tenant_databases (
    tenant_id           BIGINT UNSIGNED NOT NULL PRIMARY KEY,
    host                VARCHAR(255) NOT NULL,
    port                VARCHAR(10) NOT NULL DEFAULT '3306',
    database_name       VARCHAR(64) NOT NULL,
    username            VARCHAR(255) NOT NULL,
    encrypted_password  TEXT NOT NULL,
    status              ENUM('pending','verified','ready','error') NOT NULL DEFAULT 'pending',
    verified_at         DATETIME NULL,
    last_error          VARCHAR(500) NULL,
    created_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uq_tenant_databases_database_name (host, port, database_name),
    CONSTRAINT fk_tenant_databases_tenant
        FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
