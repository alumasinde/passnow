DROP TABLE IF EXISTS gatepass_type_gates;
ALTER TABLE gatepass_types DROP COLUMN gate_assignment_required;
ALTER TABLE gatepasses DROP FOREIGN KEY fk_gatepasses_assigned_gate;
DROP INDEX idx_gatepasses_assigned_gate ON gatepasses;
ALTER TABLE gatepasses DROP COLUMN assigned_gate_id;