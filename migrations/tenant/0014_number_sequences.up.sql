-- Compatibility migration retained so databases that already recorded 0014
-- remain consistent with the migration history. number_sequences is created
-- by 0003_visits.up.sql, where the table is first required.
SELECT 1;
