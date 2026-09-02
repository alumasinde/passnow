-- Associate login users with a tenant department. The employee table remains
-- separate because not every employee has login access.
ALTER TABLE users
    ADD COLUMN department_id BIGINT UNSIGNED NULL AFTER last_name,
    ADD KEY idx_users_department (department_id),
    ADD CONSTRAINT fk_users_department
        FOREIGN KEY (department_id) REFERENCES departments(id)
        ON DELETE SET NULL;
