-- Departments: lightweight tenant-scoped lookup for "who/where a visitor is
-- visiting" before a full Employees module exists. Kept as its own table
-- (not owned by visits) since Employees will reference it too.
CREATE TABLE departments (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tenant_id   BIGINT UNSIGNED NOT NULL,
    name        VARCHAR(120) NOT NULL,
    code        VARCHAR(30)  NOT NULL,
    active      TINYINT(1) NOT NULL DEFAULT 1,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at  DATETIME NULL,

    UNIQUE KEY uq_departments_tenant_code (tenant_id, code),
    CONSTRAINT fk_departments_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Generic concurrency-safe number sequence generator. One row per
-- (tenant_id, scope, period) — e.g. scope="visit_badge", period="2026".
-- Incremented under SELECT ... FOR UPDATE inside a transaction, so two
-- simultaneous check-ins can never receive the same number. Reused later
-- for gatepass numbering (scope="gatepass") — solved once, not per module.
CREATE TABLE number_sequences (
    tenant_id   BIGINT UNSIGNED NOT NULL,
    scope       VARCHAR(40) NOT NULL,
    period      VARCHAR(10) NOT NULL DEFAULT '', -- e.g. "2026", or "" if not period-scoped
    last_value  BIGINT UNSIGNED NOT NULL DEFAULT 0,

    PRIMARY KEY (tenant_id, scope, period),
    CONSTRAINT fk_number_sequences_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE visits (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tenant_id       BIGINT UNSIGNED NOT NULL,

    visitor_id      BIGINT UNSIGNED NOT NULL,
    visit_type_id   BIGINT UNSIGNED NULL,
    department_id   BIGINT UNSIGNED NULL,
    host_name       VARCHAR(160) NULL, -- free text until Employees module supplies a real host_employee_id FK

    purpose         VARCHAR(255) NULL,
    expected_time   DATETIME NULL,

    status          ENUM('scheduled','expected','checked_in','checked_out','cancelled','no_show','expired')
                    NOT NULL DEFAULT 'scheduled',

    -- Badge is issued at check-in only (see design note in code) — NULL
    -- until then. badge_token is the opaque value encoded in a QR/scan for
    -- fast lookup at doors without exposing badge_number sequencing.
    badge_number    VARCHAR(40)  NULL,
    badge_token     CHAR(32)     NULL,

    checked_in_at   DATETIME NULL,
    checked_in_by   BIGINT UNSIGNED NULL,
    checked_out_at  DATETIME NULL,
    checked_out_by  BIGINT UNSIGNED NULL,

    cancelled_at    DATETIME NULL,
    cancelled_by    BIGINT UNSIGNED NULL,
    cancel_reason   VARCHAR(255) NULL,

    created_by      BIGINT UNSIGNED NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at      DATETIME NULL,

    UNIQUE KEY uq_visits_tenant_badge_number (tenant_id, badge_number),
    UNIQUE KEY uq_visits_badge_token (badge_token),
    KEY idx_visits_tenant_status (tenant_id, status),
    KEY idx_visits_tenant_visitor (tenant_id, visitor_id),
    KEY idx_visits_tenant_expected (tenant_id, expected_time),

    CONSTRAINT fk_visits_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    CONSTRAINT fk_visits_visitor FOREIGN KEY (visitor_id) REFERENCES visitors(id),
    CONSTRAINT fk_visits_visit_type FOREIGN KEY (visit_type_id) REFERENCES visit_types(id),
    CONSTRAINT fk_visits_department FOREIGN KEY (department_id) REFERENCES departments(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT IGNORE INTO permissions (code, label) VALUES
    ('visits.view',     'View visits'),
    ('visits.create',   'Schedule/register a visit'),
    ('visits.checkin',  'Check in a visitor'),
    ('visits.checkout', 'Check out a visitor'),
    ('visits.cancel',   'Cancel a scheduled visit'),
    ('settings.visits', 'Manage visit settings (departments)');
