ALTER TABLE gate_devices ADD COLUMN device_secret_hash CHAR(64) NULL AFTER device_key;
ALTER TABLE gate_devices ADD UNIQUE KEY uq_gate_devices_secret_hash (device_secret_hash);
UPDATE gate_devices SET device_secret_hash = SHA2(CONCAT(device_key, ':legacy'),256) WHERE device_secret_hash IS NULL;