-- totp_secret is stored encrypted by the application. totp_last_step is the
-- time-step of the last accepted code, so a code cannot be used twice.
ALTER TABLE users
    ADD COLUMN totp_secret    text NOT NULL DEFAULT '',
    ADD COLUMN totp_enabled   boolean NOT NULL DEFAULT false,
    ADD COLUMN totp_last_step bigint NOT NULL DEFAULT 0;
