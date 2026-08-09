-- Simplification: employees is not an HR system, so drop position/phone/
-- email — a gatepass system needs "who is this person, what department,
-- can I look them up," not a full HR profile.
--
-- Keep the table itself (not merged into users or tenant_memberships):
-- most employees who appear as gatepass requesters or visit hosts never
-- log in at all, so they can't live only on tenant_memberships (which is
-- exclusively about login access). And employees is tenant-scoped data
-- about "everyone who works here" — distinct from users, which is
-- "everyone who can log in," a much smaller, cross-tenant set.
--
-- first_name/last_name become NULLABLE with a rule enforced by the CHECK
-- below: if user_id is set, the name comes from the linked users row (not
-- duplicated here — one source of truth, can't drift out of sync); if
-- user_id is NULL, this employee has no login, so the name MUST live here.
ALTER TABLE employees
    DROP COLUMN position,
    DROP COLUMN phone,
    DROP COLUMN email,
    MODIFY COLUMN first_name VARCHAR(100) NULL,
    MODIFY COLUMN last_name  VARCHAR(100) NULL,
    ADD CONSTRAINT chk_employees_name_source CHECK (
        (user_id IS NOT NULL AND first_name IS NULL AND last_name IS NULL) OR
        (user_id IS NULL AND first_name IS NOT NULL AND last_name IS NOT NULL)
    );
