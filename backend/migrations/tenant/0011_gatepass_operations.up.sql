-- Tenant-local notification outbox.
CREATE TABLE notification_outbox (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    event_key VARCHAR(180) NOT NULL,
    event_type VARCHAR(80) NOT NULL,
    entity_type VARCHAR(80) NOT NULL,
    entity_id BIGINT UNSIGNED NOT NULL,
    title VARCHAR(180) NOT NULL,
    payload JSON NULL,
    status ENUM('pending','processing','sent','failed') NOT NULL DEFAULT 'pending',
    attempts INT UNSIGNED NOT NULL DEFAULT 0,
    available_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    processed_at DATETIME NULL,
    last_error VARCHAR(500) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_notification_event_key (event_key),
    KEY idx_notification_outbox_status (status, available_at),
    KEY idx_notification_outbox_entity (entity_type, entity_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
