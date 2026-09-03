CREATE TABLE gate_devices (
 id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
 device_key CHAR(64) NOT NULL,
 name VARCHAR(120) NOT NULL,
 description TEXT NULL,
 gate_id BIGINT UNSIGNED NOT NULL,
 active TINYINT(1) NOT NULL DEFAULT 1,
 last_seen_at DATETIME NULL,
 created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
 updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
 UNIQUE KEY uq_gate_devices_key (device_key),
 KEY idx_gate_devices_gate_active (gate_id,active),
 CONSTRAINT fk_gate_devices_gate FOREIGN KEY (gate_id) REFERENCES gates(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;