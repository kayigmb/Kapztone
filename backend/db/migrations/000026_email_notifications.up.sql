-- A person can turn off the e-mail copy of notifications; in-app ones always show.
ALTER TABLE users ADD COLUMN email_notifications boolean NOT NULL DEFAULT true;

-- A notification e-mail that keeps failing is tried a few times, backing off, then dropped.
ALTER TABLE notifications
    ADD COLUMN mail_attempts int NOT NULL DEFAULT 0,
    ADD COLUMN mail_next_at  timestamptz;
