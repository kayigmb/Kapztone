-- Whether an organisation may issue certificates (see OrgRegistry.setAccredited).
ALTER TABLE organisations ADD COLUMN accredited boolean NOT NULL DEFAULT false;

-- One transform: inputs used up (in part or whole) to create output passports.
-- max_yield_ppm is the rule as it was applied, so the record stays checkable
-- after the template changes.
CREATE TABLE transforms (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organisation_id uuid NOT NULL REFERENCES organisations (id),
    commodity       text NOT NULL,
    stage           text NOT NULL,
    max_yield_ppm   bigint NOT NULL,
    data_hash       text NOT NULL,
    created_by      uuid NOT NULL REFERENCES users (id),
    tx_hash         text,
    anchored_at     timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    is_deleted      boolean NOT NULL DEFAULT false
);
CREATE INDEX transforms_organisation_id_idx ON transforms (organisation_id);
CREATE TRIGGER transforms_updated_at BEFORE UPDATE ON transforms
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- A passport made by a transform has no origin facility or period, and points
-- at the transform that made it. Its parents are that transform's inputs.
ALTER TABLE passports
    ALTER COLUMN facility_id DROP NOT NULL,
    ALTER COLUMN period DROP NOT NULL,
    ADD COLUMN transform_id uuid REFERENCES transforms (id);
CREATE INDEX passports_transform_id_idx ON passports (transform_id);

CREATE TABLE transform_inputs (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    transform_id  uuid NOT NULL REFERENCES transforms (id),
    passport_id   uuid NOT NULL REFERENCES passports (id),
    quantity_used bigint NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    is_deleted    boolean NOT NULL DEFAULT false
);
CREATE INDEX transform_inputs_transform_id_idx ON transform_inputs (transform_id);
CREATE INDEX transform_inputs_passport_id_idx ON transform_inputs (passport_id);
CREATE TRIGGER transform_inputs_updated_at BEFORE UPDATE ON transform_inputs
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- A certificate about a passport or a facility. doc_hash is the SHA-256 of
-- the file; that hash, not the file, is what goes on-chain.
CREATE TABLE attestations (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    issuer_org_id  uuid NOT NULL REFERENCES organisations (id),
    subject_type   text NOT NULL,
    subject_id     uuid NOT NULL,
    type           text NOT NULL,
    valid_from     timestamptz NOT NULL,
    valid_to       timestamptz NOT NULL,
    doc_hash       text NOT NULL,
    storage_key    text NOT NULL,
    file_name      text NOT NULL,
    content_type   text NOT NULL,
    size_bytes     bigint NOT NULL,
    ipfs_cid       text,
    status         text NOT NULL,
    revoke_reason  text,
    revoked_by     uuid REFERENCES users (id),
    revoked_at     timestamptz,
    issued_by      uuid NOT NULL REFERENCES users (id),
    tx_hash        text,
    revoke_tx_hash text,
    anchored_at    timestamptz,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),
    is_deleted     boolean NOT NULL DEFAULT false
);
CREATE INDEX attestations_subject_id_idx ON attestations (subject_id);
CREATE INDEX attestations_issuer_org_id_idx ON attestations (issuer_org_id);
CREATE TRIGGER attestations_updated_at BEFORE UPDATE ON attestations
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- A concern raised about a passport. organisation_id is who raised it.
CREATE TABLE flags (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    passport_id     uuid NOT NULL REFERENCES passports (id),
    organisation_id uuid NOT NULL REFERENCES organisations (id),
    reason          text NOT NULL,
    data_hash       text NOT NULL,
    status          text NOT NULL,
    resolution_note text,
    raised_by       uuid NOT NULL REFERENCES users (id),
    resolved_by     uuid REFERENCES users (id),
    resolved_at     timestamptz,
    tx_hash         text,
    anchored_at     timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    is_deleted      boolean NOT NULL DEFAULT false
);
CREATE INDEX flags_passport_id_idx ON flags (passport_id);
CREATE INDEX flags_organisation_id_idx ON flags (organisation_id);
CREATE TRIGGER flags_updated_at BEFORE UPDATE ON flags
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Something an organisation should know about. emailed_at is set once the
-- e-mail copy has been sent.
CREATE TABLE notifications (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organisation_id uuid NOT NULL REFERENCES organisations (id),
    kind            text NOT NULL,
    subject_id      uuid NOT NULL,
    message         text NOT NULL,
    read_at         timestamptz,
    emailed_at      timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    is_deleted      boolean NOT NULL DEFAULT false
);
CREATE INDEX notifications_organisation_id_idx ON notifications (organisation_id, created_at);
CREATE INDEX notifications_unsent_idx ON notifications (created_at) WHERE emailed_at IS NULL;
CREATE TRIGGER notifications_updated_at BEFORE UPDATE ON notifications
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
