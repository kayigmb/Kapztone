-- Every successful change made through the API, one row per request. Rows
-- form a hash chain: each hash covers the row and the previous row's hash, so
-- editing or removing a row breaks every hash after it (GET /audit/verify).
-- No foreign keys: the log must outlive, and never block, what it records.
CREATE TABLE audit_log (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    seq             bigint NOT NULL UNIQUE,
    actor_user_id   uuid,
    organisation_id uuid,
    method          text NOT NULL,
    route           text NOT NULL,
    subject_id      text,
    status          integer NOT NULL,
    prev_hash       text NOT NULL,
    hash            text NOT NULL,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    is_deleted      boolean NOT NULL DEFAULT false
);
CREATE INDEX audit_log_organisation_id_idx ON audit_log (organisation_id);
CREATE INDEX audit_log_actor_user_id_idx ON audit_log (actor_user_id);
CREATE TRIGGER audit_log_updated_at BEFORE UPDATE ON audit_log
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
