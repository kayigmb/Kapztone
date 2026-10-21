-- A new sign-up confirms its email before it can sign in. Accounts that
-- exist already count as confirmed.
ALTER TABLE users ADD COLUMN email_verified_at timestamptz;
UPDATE users SET email_verified_at = now();
