-- Batch 1: Gatepass domain hardening.
-- Adds explicit lifecycle states and a tenant-configurable returnability policy.
-- Existing rows remain valid: their policy defaults to OPTIONAL.

ALTER TABLE gatepass_types
    ADD COLUMN returnability_policy ENUM('optional','required','not_allowed')
    NOT NULL DEFAULT 'optional'
    AFTER is_returnable_default;

ALTER TABLE gatepasses
    MODIFY COLUMN status ENUM(
        'draft',
        'submitted',
        'pending_approval',
        'approved',
        'rejected',
        'cancelled',
        'expired',
        'checked_out',
        'awaiting_return',
        'partially_returned',
        'return_overdue',
        'checked_in',
        'completed'
    ) NOT NULL DEFAULT 'pending_approval';

CREATE INDEX idx_gatepasses_tenant_return_due
    ON gatepasses (tenant_id, status, expected_return_at);

-- Data safety: existing non-returnable passes must not carry a return date.
UPDATE gatepasses
SET expected_return_at = NULL
WHERE is_returnable = 0;

-- Normalize existing types so the policy agrees with the old default flag.
UPDATE gatepass_types
SET returnability_policy = CASE
    WHEN is_returnable_default = 1 THEN 'required'
    ELSE 'optional'
END
WHERE returnability_policy = 'optional';

-- The old is_returnable_default flag remains as the default value for OPTIONAL
-- types. REQUIRED/NOT_ALLOWED are enforced by the application policy.
