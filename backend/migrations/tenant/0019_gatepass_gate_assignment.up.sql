ALTER TABLE gatepasses ADD COLUMN assigned_gate_id BIGINT UNSIGNED NULL AFTER gatepass_type_id;
ALTER TABLE gatepasses ADD CONSTRAINT fk_gatepasses_assigned_gate FOREIGN KEY (assigned_gate_id) REFERENCES gates(id);
CREATE INDEX idx_gatepasses_assigned_gate ON gatepasses(assigned_gate_id);

ALTER TABLE gatepass_types ADD COLUMN gate_assignment_required TINYINT(1) NOT NULL DEFAULT 0 AFTER direction;

CREATE TABLE gatepass_type_gates (
    gatepass_type_id BIGINT UNSIGNED NOT NULL,
    gate_id BIGINT UNSIGNED NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (gatepass_type_id, gate_id),
    CONSTRAINT fk_gatepass_type_gates_type FOREIGN KEY (gatepass_type_id) REFERENCES gatepass_types(id) ON DELETE CASCADE,
    CONSTRAINT fk_gatepass_type_gates_gate FOREIGN KEY (gate_id) REFERENCES gates(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;