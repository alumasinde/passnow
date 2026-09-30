ALTER TABLE gate_devices DROP INDEX uq_gate_devices_secret_hash;
ALTER TABLE gate_devices DROP COLUMN device_secret_hash;