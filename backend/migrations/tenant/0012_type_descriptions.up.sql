-- Settings type descriptions.
-- These are tenant-local databases. Descriptions are optional admin-facing
-- context and are returned by the same type APIs used by the Settings UI.

ALTER TABLE gatepass_types
    ADD COLUMN description VARCHAR(255) NULL AFTER code;

ALTER TABLE visit_types
    ADD COLUMN description VARCHAR(255) NULL AFTER code;
