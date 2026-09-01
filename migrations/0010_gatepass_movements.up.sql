-- Batch 3: Physical gate movement ledger.
-- Gatepass header timestamps remain as convenient summary fields, while the
-- movement tables are the authoritative audit trail for physical movement.

CREATE TABLE IF NOT EXISTS gatepass_movements (
    id             BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tenant_id      BIGINT UNSIGNED NOT NULL,
    gatepass_id    BIGINT UNSIGNED NOT NULL,
    type           ENUM('checkout','checkin') NOT NULL,
    actor_user_id  BIGINT UNSIGNED NOT NULL,
    gate_name      VARCHAR(120) NULL,
    notes          VARCHAR(500) NULL,
    occurred_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    KEY idx_gp_movements_tenant_gatepass_time (tenant_id, gatepass_id, occurred_at, id),
    KEY idx_gp_movements_tenant_type_time (tenant_id, type, occurred_at),

    CONSTRAINT fk_gpm_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    CONSTRAINT fk_gpm_gatepass FOREIGN KEY (gatepass_id) REFERENCES gatepasses(id) ON DELETE CASCADE,
    CONSTRAINT fk_gpm_actor FOREIGN KEY (actor_user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS gatepass_movement_items (
    id               BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tenant_id        BIGINT UNSIGNED NOT NULL,
    movement_id      BIGINT UNSIGNED NOT NULL,
    gatepass_item_id BIGINT UNSIGNED NOT NULL,
    quantity         DECIMAL(10,2) NOT NULL,
    outcome          ENUM('released','returned','damaged','lost') NOT NULL DEFAULT 'returned',
    `condition`      VARCHAR(80) NULL,
    notes            VARCHAR(500) NULL,
    created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    KEY idx_gpm_items_movement (tenant_id, movement_id),
    KEY idx_gpm_items_gatepass_item (tenant_id, gatepass_item_id),

    CONSTRAINT fk_gpmi_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    CONSTRAINT fk_gpmi_movement FOREIGN KEY (movement_id) REFERENCES gatepass_movements(id) ON DELETE CASCADE,
    CONSTRAINT fk_gpmi_gatepass_item FOREIGN KEY (gatepass_item_id) REFERENCES gatepass_items(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
