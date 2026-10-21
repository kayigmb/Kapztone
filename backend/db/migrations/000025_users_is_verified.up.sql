-- Whether the email is confirmed, as a flag sign-in checks. Derived from
-- email_verified_at, so the two can never disagree.
ALTER TABLE users ADD COLUMN is_verified boolean GENERATED ALWAYS AS (email_verified_at IS NOT NULL) STORED;
