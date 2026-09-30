ALTER TABLE visits
  DROP INDEX idx_visits_entry_source,
  DROP INDEX idx_visits_expected_departure,
  DROP COLUMN arrived_at,
  DROP COLUMN expected_departure_at,
  DROP COLUMN entry_source;