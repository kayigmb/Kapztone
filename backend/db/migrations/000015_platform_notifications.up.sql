-- A notification with no organisation is for the platform (its admins).
ALTER TABLE notifications ALTER COLUMN organisation_id DROP NOT NULL;
