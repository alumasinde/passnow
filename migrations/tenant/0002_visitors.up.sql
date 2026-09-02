-- Tenant-local key/value settings store. Used for tenant administration
-- toggles like "can visitors be pre-registered" without a schema change
-- per feature flag. value is JSON so a setting can hold bool/string/number/
-- object as needed; callers agree on the shape per key.
CREATE TABLE tenant_settings (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    setting_key VARCHAR(120) NOT NULL,      -- e.g. "visitors.allow_pre_registration"
    value       JSON NOT NULL,
    updated_by  BIGINT UNSIGNED NULL,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uq_setting (setting_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Configurable ID document types. (National ID, Passport, ...).
CREATE TABLE id_types (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name            VARCHAR(100) NOT NULL,
    code            VARCHAR(30)  NOT NULL,   -- short code, e.g. "NATID", "PASSPORT"
    requires_number TINYINT(1) NOT NULL DEFAULT 1,
    active          TINYINT(1) NOT NULL DEFAULT 1,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at      DATETIME NULL,

    UNIQUE KEY uq_id_types_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Configurable visit purposes/types. (Business Meeting, Delivery,
-- Interview, Maintenance, ...). Full visit-lifecycle usage comes with the
-- Visits module; the lookup itself is built now since it's shared config.
CREATE TABLE visit_types (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    code        VARCHAR(30)  NOT NULL,
    active      TINYINT(1) NOT NULL DEFAULT 1,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at  DATETIME NULL,

    UNIQUE KEY uq_visit_types_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Organizations visitors belong to/represent.
CREATE TABLE visitor_companies (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name        VARCHAR(160) NOT NULL,
    phone       VARCHAR(30)  NULL,
    email       VARCHAR(255) NULL,
    address     VARCHAR(255) NULL,
    active      TINYINT(1) NOT NULL DEFAULT 1,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at  DATETIME NULL,

    UNIQUE KEY uq_visitor_companies_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE visitors (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,

    first_name      VARCHAR(100) NOT NULL,
    last_name       VARCHAR(100) NOT NULL,

    id_type_id      BIGINT UNSIGNED NOT NULL,
    id_number       VARCHAR(60)  NULL,   -- nullable: some id_types don't require one

    company_id      BIGINT UNSIGNED NULL,

    phone           VARCHAR(30)  NULL,
    email           VARCHAR(255) NULL,
    photo_ref       VARCHAR(255) NULL,   -- storage path/URL reference, not the file itself
    notes           VARCHAR(500) NULL,

    source          ENUM('walk_in','pre_registered') NOT NULL DEFAULT 'walk_in',
    status          ENUM('active','blacklisted') NOT NULL DEFAULT 'active',
    blacklist_reason VARCHAR(255) NULL,

    created_by      BIGINT UNSIGNED NULL,
    updated_by      BIGINT UNSIGNED NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at      DATETIME NULL,

    -- Prevents the same ID document being registered twice within this database. NULL id_number rows are not constrained by this (MySQL
    -- treats NULLs as distinct), which is correct for id_types that don't
    -- require a number.
    UNIQUE KEY uq_visitors_idtype_idnumber (id_type_id, id_number),
    KEY idx_visitors_name (last_name, first_name),
    KEY idx_visitors_company (company_id),
    KEY idx_visitors_status (status),
    CONSTRAINT fk_visitors_id_type FOREIGN KEY (id_type_id) REFERENCES id_types(id),
    CONSTRAINT fk_visitors_company FOREIGN KEY (company_id) REFERENCES visitor_companies(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Permission codes this module introduces. INSERT IGNORE so re-running
-- migrations (or applying out of order in dev) never errors on duplicates.
INSERT IGNORE INTO permissions (code, label) VALUES
    ('visitors.view',    'View visitors'),
    ('visitors.create',  'Register visitors'),
    ('visitors.update',  'Edit visitors'),
    ('settings.visitors','Manage visitor settings (ID types, visit types, companies, pre-registration)');
