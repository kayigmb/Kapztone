-- Which certificate types an organisation may issue, and for how long (see
-- OrgRegistry.setAccreditation). cert_type is an evidence key from a template.
CREATE TABLE accreditations (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organisation_id uuid NOT NULL REFERENCES organisations (id),
    cert_type       text NOT NULL,
    valid_from      timestamptz NOT NULL,
    valid_to        timestamptz NOT NULL,
    tx_hash         text,
    anchored_at     timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    is_deleted      boolean NOT NULL DEFAULT false
);
CREATE UNIQUE INDEX accreditations_org_type_key ON accreditations (organisation_id, cert_type) WHERE NOT is_deleted;
CREATE TRIGGER accreditations_updated_at BEFORE UPDATE ON accreditations
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE organisations DROP COLUMN accredited;
