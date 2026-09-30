-- Keep the platform users schema compatible with the shared users repository.
-- Tenant users received this column in the tenant migration set; platform
-- administrators authenticate from the platform users table and need it too.
ALTER TABLE users
    ADD COLUMN must_change_password TINYINT(1) NOT NULL DEFAULT 0
    AFTER status;
