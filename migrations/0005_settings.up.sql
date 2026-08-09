-- Employees: the org directory (department, position, employee number),
-- separate from `users` (login accounts). An employee MAY be linked to a
-- user account (user_id) if they need to log in — e.g. to request their
-- own gatepasses — but plenty of employees (no system access) can exist
-- without one, and plenty of logged-in users (contractors doing data
-- entry) may not need an employee record at all.
CREATE TABLE employees (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tenant_id       BIGINT UNSIGNED NOT NULL,
    employee_number VARCHAR(40)  NOT NULL,
    first_name      VARCHAR(100) NOT NULL,
    last_name       VARCHAR(100) NOT NULL,
    department_id   BIGINT UNSIGNED NULL,
    position        VARCHAR(120) NULL,
    phone           VARCHAR(30)  NULL,
    email           VARCHAR(255) NULL,
    user_id         BIGINT UNSIGNED NULL, -- optional link to a login account
    status          ENUM('active','inactive') NOT NULL DEFAULT 'active',

    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at      DATETIME NULL,

    UNIQUE KEY uq_employees_tenant_number (tenant_id, employee_number),
    KEY idx_employees_tenant_department (tenant_id, department_id),
    KEY idx_employees_tenant_user (tenant_id, user_id),
    CONSTRAINT fk_employees_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    CONSTRAINT fk_employees_department FOREIGN KEY (department_id) REFERENCES departments(id),
    CONSTRAINT fk_employees_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT IGNORE INTO permissions (code, label) VALUES
    ('employees.view',    'View employees'),
    ('employees.create',  'Create employees'),
    ('employees.update',  'Edit employees'),
    ('settings.users',       'Manage users and tenant memberships'),
    ('settings.roles',       'Manage roles'),
    ('settings.permissions', 'Assign permissions to roles');
