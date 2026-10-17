-- Signing PIN: asked again on every signing action (issue, handover, transform,
-- certificate, flag resolution). Stored as a bcrypt hash; NULL means not set.
ALTER TABLE users ADD COLUMN pin_hash text;

-- One-time codes for signing in when the authenticator app is lost. Only the
-- SHA-256 of each code is kept; a used code keeps its row with used_at set.
CREATE TABLE recovery_codes (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid NOT NULL REFERENCES users (id),
    code_hash  text NOT NULL,
    used_at    timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    is_deleted boolean NOT NULL DEFAULT false
);
CREATE INDEX recovery_codes_user_idx ON recovery_codes (user_id) WHERE NOT is_deleted;
CREATE TRIGGER recovery_codes_updated_at BEFORE UPDATE ON recovery_codes
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- An invitation to join an organisation with a role. Only the SHA-256 of the
-- emailed token is kept; it works once, until expires_at.
CREATE TABLE invites (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organisation_id uuid NOT NULL REFERENCES organisations (id),
    email           citext NOT NULL,
    role            text NOT NULL,
    token_hash      text NOT NULL,
    expires_at      timestamptz NOT NULL,
    accepted_at     timestamptz,
    invited_by      uuid NOT NULL REFERENCES users (id),
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    is_deleted      boolean NOT NULL DEFAULT false
);
CREATE UNIQUE INDEX invites_token_key ON invites (token_hash) WHERE NOT is_deleted;
CREATE TRIGGER invites_updated_at BEFORE UPDATE ON invites
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
