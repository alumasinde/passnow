CREATE TABLE visit_movements (
 id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
 visit_id BIGINT UNSIGNED NOT NULL,
 movement_type VARCHAR(20) NOT NULL,
 gate_id BIGINT UNSIGNED NOT NULL,
 device_id BIGINT UNSIGNED NULL,
 actor_user_id BIGINT UNSIGNED NOT NULL,
 notes TEXT NULL,
 occurred_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
 created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
 CONSTRAINT fk_visit_movements_visit FOREIGN KEY (visit_id) REFERENCES visits(id),
 CONSTRAINT fk_visit_movements_gate FOREIGN KEY (gate_id) REFERENCES gates(id),
 CONSTRAINT fk_visit_movements_device FOREIGN KEY (device_id) REFERENCES gate_devices(id),
 KEY idx_visit_movements_visit (visit_id),
 KEY idx_visit_movements_gate (gate_id),
 KEY idx_visit_movements_occurred (occurred_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;