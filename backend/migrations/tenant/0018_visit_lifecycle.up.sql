ALTER TABLE visits
  ADD COLUMN entry_source VARCHAR(40) NOT NULL DEFAULT 'pre_registered' AFTER visitor_id,
  ADD COLUMN expected_departure_at DATETIME NULL AFTER expected_time,
  ADD COLUMN arrived_at DATETIME NULL AFTER expected_departure_at;

ALTER TABLE visits
  ADD KEY idx_visits_entry_source (entry_source),
  ADD KEY idx_visits_expected_departure (expected_departure_at);

UPDATE visits
SET entry_source = 'pre_registered'
WHERE entry_source = '';