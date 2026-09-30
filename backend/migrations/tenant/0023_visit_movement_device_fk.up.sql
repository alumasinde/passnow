ALTER TABLE visit_movements
 ADD CONSTRAINT fk_visit_movements_device
 FOREIGN KEY (device_id) REFERENCES gate_devices(id);