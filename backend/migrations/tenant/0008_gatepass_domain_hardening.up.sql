-- Gatepass lifecycle hardening for one tenant database.
ALTER TABLE gatepass_types
    ADD COLUMN returnability_policy ENUM('optional','required','not_allowed')
    NOT NULL DEFAULT 'optional'
    AFTER is_returnable_default;

ALTER TABLE gatepasses
    MODIFY COLUMN status ENUM(
        'draft','submitted','pending_approval','approved','rejected','cancelled',
        'expired','checked_out','awaiting_return','partially_returned',
        'return_overdue','checked_in','completed'
    ) NOT NULL DEFAULT 'pending_approval';

CREATE INDEX idx_gatepasses_return_due
    ON gatepasses (status, expected_return_at);

UPDATE gatepasses
SET expected_return_at = NULL
WHERE is_returnable = 0;

UPDATE gatepass_types
SET returnability_policy = CASE
    WHEN is_returnable_default = 1 THEN 'required'
    ELSE 'optional'
END
WHERE returnability_policy = 'optional';
