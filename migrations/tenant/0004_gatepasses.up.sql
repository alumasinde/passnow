-- Approval workflow TEMPLATES. A gatepass_type points at one of these;
-- when a gatepass is created, its steps are SNAPSHOTTED onto the gatepass
-- (gatepass_approvals) so editing a template later never changes an
-- approval already in progress.
CREATE TABLE approval_workflows (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name        VARCHAR(120) NOT NULL,
    active      TINYINT(1) NOT NULL DEFAULT 1,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at  DATETIME NULL,

    UNIQUE KEY uq_workflows_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE approval_workflow_steps (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    workflow_id     BIGINT UNSIGNED NOT NULL,
    step_order      SMALLINT UNSIGNED NOT NULL,
    label           VARCHAR(80) NOT NULL,      -- e.g. "HOD", "Security Manager", "General Manager"
    approver_type   ENUM('role','specific_user') NOT NULL DEFAULT 'role',
    role_id         BIGINT UNSIGNED NULL,
    user_id         BIGINT UNSIGNED NULL,
    required        TINYINT(1) NOT NULL DEFAULT 1,

    UNIQUE KEY uq_workflow_step_order (workflow_id, step_order),
    CONSTRAINT fk_wfsteps_workflow FOREIGN KEY (workflow_id) REFERENCES approval_workflows(id) ON DELETE CASCADE,
    CONSTRAINT fk_wfsteps_role FOREIGN KEY (role_id) REFERENCES roles(id),
    CONSTRAINT fk_wfsteps_user FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT chk_wfsteps_approver CHECK (
        (approver_type = 'role' AND role_id IS NOT NULL AND user_id IS NULL) OR
        (approver_type = 'specific_user' AND user_id IS NOT NULL AND role_id IS NULL)
    )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Configurable gatepass types (Visitor Gatepass, Material Out, Equipment
-- Removal, ...). direction + is_returnable_default drive which gate
-- action(s) are legal and whether a return is expected — see the Go state
-- machine in internal/gatepasses for the full rules.
CREATE TABLE gatepass_types (
    id                   BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name                 VARCHAR(120) NOT NULL,
    code                 VARCHAR(30)  NOT NULL,
    direction            ENUM('in','out','both') NOT NULL DEFAULT 'out',
    is_returnable_default TINYINT(1) NOT NULL DEFAULT 0,
    requires_items       TINYINT(1) NOT NULL DEFAULT 0,

    -- requires_approval is an ADMIN-set mandate: when true, a gatepass of
    -- this type can NEVER be created without going through workflow_id's
    -- steps, regardless of what the client sends (see service.go — the
    -- request's "needs approval" checkbox is only honored as an opt-IN
    -- extra, never as a bypass).
    requires_approval    TINYINT(1) NOT NULL DEFAULT 0,
    workflow_id          BIGINT UNSIGNED NULL,

    active               TINYINT(1) NOT NULL DEFAULT 1,
    created_at           DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at           DATETIME NULL,

    UNIQUE KEY uq_gatepass_types_code (code),
    CONSTRAINT fk_gptypes_workflow FOREIGN KEY (workflow_id) REFERENCES approval_workflows(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE gatepasses (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,

    gatepass_type_id BIGINT UNSIGNED NOT NULL,
    pass_number     VARCHAR(50) NOT NULL,

    department_id   BIGINT UNSIGNED NULL,

    requester_type      ENUM('employee','visitor') NOT NULL,
    requester_user_id    BIGINT UNSIGNED NULL, -- employee path: the logged-in user
    requester_visitor_id BIGINT UNSIGNED NULL, -- visitor path: looked up by ID number

    visit_id        BIGINT UNSIGNED NULL, -- optional link to an existing visit

    purpose         VARCHAR(255) NULL,

    is_returnable    TINYINT(1) NOT NULL DEFAULT 0,
    expected_return_at DATETIME NULL,

    -- requires_approval here is the ACTUAL resolved value for this
    -- specific gatepass (type mandate OR requester opt-in) — see
    -- service.go. workflow_id is which template was snapshotted, kept for
    -- reference even though gatepass_approvals holds the real steps.
    requires_approval TINYINT(1) NOT NULL DEFAULT 0,
    workflow_id       BIGINT UNSIGNED NULL,

    status          ENUM('pending_approval','approved','rejected','cancelled','checked_out','checked_in')
                    NOT NULL DEFAULT 'pending_approval',

    qr_token        CHAR(32) NOT NULL,

    checked_out_at  DATETIME NULL,
    checked_out_by  BIGINT UNSIGNED NULL,
    checked_in_at   DATETIME NULL,
    checked_in_by   BIGINT UNSIGNED NULL,

    cancelled_at    DATETIME NULL,
    cancelled_by    BIGINT UNSIGNED NULL,
    cancel_reason   VARCHAR(255) NULL,

    issued_by       BIGINT UNSIGNED NULL, -- who approved the final step / auto-issued it
    issued_at       DATETIME NULL,

    created_by      BIGINT UNSIGNED NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at      DATETIME NULL,

    UNIQUE KEY uq_gatepasses_number (pass_number),
    UNIQUE KEY uq_gatepasses_qr_token (qr_token),
    KEY idx_gatepasses_status (status),
    KEY idx_gatepasses_requester_user (requester_user_id),
    KEY idx_gatepasses_requester_visitor (requester_visitor_id),
    KEY idx_gatepasses_visit (visit_id),
    CONSTRAINT fk_gp_type FOREIGN KEY (gatepass_type_id) REFERENCES gatepass_types(id),
    CONSTRAINT fk_gp_department FOREIGN KEY (department_id) REFERENCES departments(id),
    CONSTRAINT fk_gp_requester_user FOREIGN KEY (requester_user_id) REFERENCES users(id),
    CONSTRAINT fk_gp_requester_visitor FOREIGN KEY (requester_visitor_id) REFERENCES visitors(id),
    CONSTRAINT fk_gp_visit FOREIGN KEY (visit_id) REFERENCES visits(id),
    CONSTRAINT fk_gp_workflow FOREIGN KEY (workflow_id) REFERENCES approval_workflows(id),
    CONSTRAINT chk_gp_requester CHECK (
        (requester_type = 'employee' AND requester_user_id IS NOT NULL AND requester_visitor_id IS NULL) OR
        (requester_type = 'visitor' AND requester_visitor_id IS NOT NULL AND requester_user_id IS NULL)
    )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Snapshotted approval steps for one specific gatepass. See table comment
-- on approval_workflow_steps for why this is a copy, not a live join.
CREATE TABLE gatepass_approvals (
    id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    gatepass_id   BIGINT UNSIGNED NOT NULL,
    step_order    SMALLINT UNSIGNED NOT NULL,
    label         VARCHAR(80) NOT NULL,
    approver_type ENUM('role','specific_user') NOT NULL,
    role_id       BIGINT UNSIGNED NULL,
    user_id       BIGINT UNSIGNED NULL,
    required      TINYINT(1) NOT NULL DEFAULT 1,

    status        ENUM('pending','approved','rejected','skipped') NOT NULL DEFAULT 'pending',
    acted_by      BIGINT UNSIGNED NULL,
    acted_at      DATETIME NULL,
    comments      VARCHAR(500) NULL,

    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY uq_gp_approval_step (gatepass_id, step_order),
    KEY idx_gp_approvals_gatepass (gatepass_id),
    CONSTRAINT fk_gpapp_gatepass FOREIGN KEY (gatepass_id) REFERENCES gatepasses(id) ON DELETE CASCADE,
    CONSTRAINT fk_gpapp_role FOREIGN KEY (role_id) REFERENCES roles(id),
    CONSTRAINT fk_gpapp_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE gatepass_items (
    id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    gatepass_id   BIGINT UNSIGNED NOT NULL,

    name          VARCHAR(160) NOT NULL,
    description   VARCHAR(255) NULL,
    category      VARCHAR(80)  NULL,
    quantity      DECIMAL(10,2) NOT NULL DEFAULT 1,
    unit          VARCHAR(20)  NULL,
    serial_number VARCHAR(80)  NULL,
    asset_number  VARCHAR(80)  NULL,
    item_condition VARCHAR(80) NULL,
    direction     ENUM('entering','leaving','returning') NOT NULL,

    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    KEY idx_gp_items_gatepass (gatepass_id),
    CONSTRAINT fk_gpitems_gatepass FOREIGN KEY (gatepass_id) REFERENCES gatepasses(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT IGNORE INTO permissions (code, label) VALUES
    ('gatepasses.view',    'View gatepasses'),
    ('gatepasses.create',  'Create gatepasses'),
    ('gatepasses.update',  'Edit gatepasses'),
    ('gatepasses.approve', 'Approve gatepass approval steps'),
    ('gatepasses.reject',  'Reject gatepass approval steps'),
    ('gatepasses.cancel',  'Cancel gatepasses'),
    ('gatepasses.issue',   'Check items/visitors out at the gate'),
    ('gatepasses.verify',  'Check items/visitors in at the gate'),
    ('settings.gatepass',  'Manage gatepass types'),
    ('settings.approvals', 'Manage approval workflows');
