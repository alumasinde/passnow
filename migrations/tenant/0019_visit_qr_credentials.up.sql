ALTER TABLE visits
  ADD COLUMN qr_token CHAR(32) NULL AFTER badge_token,
  ADD COLUMN qr_issued_at DATETIME NULL AFTER qr_token,
  ADD COLUMN qr_invalidated_at DATETIME NULL AFTER qr_issued_at,
  ADD UNIQUE KEY uq_visits_qr_token (qr_token),
  ADD KEY idx_visits_qr_invalidated (qr_invalidated_at);