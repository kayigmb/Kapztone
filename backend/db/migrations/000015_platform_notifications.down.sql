DELETE FROM notifications WHERE organisation_id IS NULL;
ALTER TABLE notifications ALTER COLUMN organisation_id SET NOT NULL;
