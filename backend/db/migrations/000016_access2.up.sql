-- Every sign-in token carries the user's token version; bumping it (password
-- change or reset, "sign out other devices", lock) voids every token issued before.
ALTER TABLE users ADD COLUMN token_version integer NOT NULL DEFAULT 0;

-- Hashes of the passwords a user has set, so a new one can't repeat a recent one.
CREATE TABLE password_history (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       uuid NOT NULL REFERENCES users (id),
    password_hash text NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    is_deleted    boolean NOT NULL DEFAULT false
);
CREATE INDEX password_history_user_idx ON password_history (user_id, created_at DESC) WHERE NOT is_deleted;
CREATE TRIGGER password_history_updated_at BEFORE UPDATE ON password_history
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Emailed one-use links: a password reset ("reset", 30 minutes) or the
-- "Lock my account" link sent after a password change ("lock", 7 days).
-- Only the SHA-256 of the token is kept.
CREATE TABLE password_resets (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid NOT NULL REFERENCES users (id),
    kind       text NOT NULL,
    token_hash text NOT NULL,
    expires_at timestamptz NOT NULL,
    used_at    timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    is_deleted boolean NOT NULL DEFAULT false
);
CREATE UNIQUE INDEX password_resets_token_key ON password_resets (token_hash) WHERE NOT is_deleted;
CREATE TRIGGER password_resets_updated_at BEFORE UPDATE ON password_resets
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
