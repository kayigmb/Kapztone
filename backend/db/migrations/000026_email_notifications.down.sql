ALTER TABLE notifications DROP COLUMN mail_attempts, DROP COLUMN mail_next_at;
ALTER TABLE users DROP COLUMN email_notifications;
