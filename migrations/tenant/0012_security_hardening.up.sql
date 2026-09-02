-- Database-per-tenant security hardening.
-- Refresh tokens are tenant-bound by the database connection; no tenant_id is stored.
CREATE INDEX idx_refresh_user_revoked
    ON refresh_tokens (user_id, revoked_at);
