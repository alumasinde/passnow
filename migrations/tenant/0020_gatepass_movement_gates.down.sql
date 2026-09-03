ALTER TABLE gatepass_movements DROP FOREIGN KEY fk_gatepass_movements_gate;
DROP INDEX idx_gatepass_movements_gate ON gatepass_movements;
ALTER TABLE gatepass_movements DROP COLUMN gate_id;