-- Database-per-tenant architecture.
-- Tenant isolation is enforced by the database boundary. Ordinary local foreign
-- keys provide referential integrity; tenant_id composite keys are not required.
SELECT 1;
