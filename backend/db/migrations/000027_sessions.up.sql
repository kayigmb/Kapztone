-- One sign-in in one browser, shared by every address a handoff carries it to:
-- signing out anywhere ends it everywhere in that browser.
CREATE TABLE sessions (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid NOT NULL REFERENCES users (id),
    ended_at   timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    is_deleted boolean NOT NULL DEFAULT false
);

-- A handoff code carries the session it continues.
ALTER TABLE password_resets ADD COLUMN session_id uuid REFERENCES sessions (id);
