-- Batch 5: Security hardening.
-- Bind refresh tokens to the tenant in which they were issued. This closes
-- a cross-tenant refresh-token replay path for users who belong to multiple
-- tenants.

ALTER TABLE refresh_tokens
    ADD COLUMN tenant_id BIGINT UNSIGNED NULL AFTER user_id;

-- Only migrate tokens for users who belong to exactly ONE tenant.
-- A multi-tenant user's old token has no trustworthy tenant binding.
UPDATE refresh_tokens rt
JOIN (
    SELECT user_id, MIN(tenant_id) AS tenant_id
    FROM tenant_memberships
    WHERE status = 'active'
    GROUP BY user_id
    HAVING COUNT(*) = 1
) tm ON tm.user_id = rt.user_id
SET rt.tenant_id = tm.tenant_id
WHERE rt.tenant_id IS NULL;

-- Ambiguous legacy tokens (including users with no active membership) cannot
-- safely be assigned to one tenant. Revoke them before making the binding
-- mandatory.
UPDATE refresh_tokens rt
SET revoked_at = COALESCE(revoked_at, CURRENT_TIMESTAMP)
WHERE rt.tenant_id IS NULL;

ALTER TABLE refresh_tokens
    MODIFY COLUMN tenant_id BIGINT UNSIGNED NOT NULL;

ALTER TABLE refresh_tokens
    ADD CONSTRAINT fk_refresh_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id);

CREATE INDEX idx_refresh_tenant_user ON refresh_tokens (tenant_id, user_id, revoked_at);

-- Tenant-owned roles must never point at a role from a different tenant.
-- The application already validates this; these indexes support the checks.
CREATE INDEX idx_membership_tenant_role_user
    ON tenant_memberships (tenant_id, role_id, user_id, status);
