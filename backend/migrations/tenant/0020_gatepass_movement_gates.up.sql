ALTER TABLE gatepass_movements ADD COLUMN gate_id BIGINT UNSIGNED NULL AFTER actor_user_id;
ALTER TABLE gatepass_movements ADD CONSTRAINT fk_gatepass_movements_gate FOREIGN KEY (gate_id) REFERENCES gates(id);
CREATE INDEX idx_gatepass_movements_gate ON gatepass_movements(gate_id);