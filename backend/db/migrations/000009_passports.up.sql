-- A digital product passport for one lot. organisation_id is the current holder.
-- Quantities are whole numbers in `unit`. tx_hash and anchored_at are filled
-- once the record is written to the chain.
CREATE TABLE passports (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organisation_id uuid NOT NULL REFERENCES organisations (id),
    facility_id     uuid NOT NULL REFERENCES facilities (id),
    template_id     uuid NOT NULL REFERENCES commodity_templates (id),
    commodity       text NOT NULL,
    unit            text NOT NULL,
    quantity        bigint NOT NULL,
    remaining       bigint NOT NULL,
    period          bigint NOT NULL,
    status          text NOT NULL,
    data_hash       text NOT NULL,
    handover_count  bigint NOT NULL DEFAULT 0,
    issued_by       uuid NOT NULL REFERENCES users (id),
    tx_hash         text,
    anchored_at     timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    is_deleted      boolean NOT NULL DEFAULT false
);
CREATE INDEX passports_organisation_id_idx ON passports (organisation_id);
CREATE INDEX passports_facility_period_idx ON passports (facility_id, period);
CREATE TRIGGER passports_updated_at BEFORE UPDATE ON passports
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- A custody handover. It carries the sender's signature from the start and
-- the receiver's once accepted; only then does the passport change holder.
CREATE TABLE handovers (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    passport_id   uuid NOT NULL REFERENCES passports (id),
    from_org      uuid NOT NULL REFERENCES organisations (id),
    to_org        uuid NOT NULL REFERENCES organisations (id),
    quantity      bigint NOT NULL,
    note          text NOT NULL,
    data_hash     text NOT NULL,
    nonce         bigint NOT NULL,
    deadline      timestamptz NOT NULL,
    sig_from      text NOT NULL,
    sig_to        text,
    status        text NOT NULL,
    reject_reason text,
    created_by    uuid NOT NULL REFERENCES users (id),
    responded_by  uuid REFERENCES users (id),
    responded_at  timestamptz,
    tx_hash       text,
    anchored_at   timestamptz,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    is_deleted    boolean NOT NULL DEFAULT false
);
CREATE INDEX handovers_passport_id_idx ON handovers (passport_id);
CREATE INDEX handovers_from_org_idx ON handovers (from_org);
CREATE INDEX handovers_to_org_idx ON handovers (to_org);
CREATE TRIGGER handovers_updated_at BEFORE UPDATE ON handovers
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- One step of a passport's journey. ref_id is the row the step came from
-- (the passport for an issue, the handover for a handover).
CREATE TABLE trace_events (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    passport_id uuid NOT NULL REFERENCES passports (id),
    ref_id      uuid NOT NULL,
    kind        text NOT NULL,
    from_org    uuid REFERENCES organisations (id),
    to_org      uuid REFERENCES organisations (id),
    quantity    bigint NOT NULL,
    data_hash   text NOT NULL,
    tx_hash     text,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    is_deleted  boolean NOT NULL DEFAULT false
);
CREATE INDEX trace_events_passport_id_idx ON trace_events (passport_id);
CREATE INDEX trace_events_ref_id_idx ON trace_events (ref_id);
CREATE TRIGGER trace_events_updated_at BEFORE UPDATE ON trace_events
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
