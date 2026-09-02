-- Tenant-local numbering sequences. Isolation is provided by the database
-- connection, so tenant_id is intentionally not stored here.
CREATE TABLE IF NOT EXISTS number_sequences (
    scope      VARCHAR(64) NOT NULL,
    period     VARCHAR(32) NOT NULL DEFAULT '',
    last_value BIGINT UNSIGNED NOT NULL DEFAULT 0,
    PRIMARY KEY (scope, period)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
