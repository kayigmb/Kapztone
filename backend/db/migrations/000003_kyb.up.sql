ALTER TABLE organisations
    ADD COLUMN reviewed_by   uuid REFERENCES users (id),
    ADD COLUMN reviewed_at   timestamptz,
    ADD COLUMN review_reason text;

-- KYB evidence. The file itself lives on disk under storage_key; sha256 proves
-- the stored file is the one that was reviewed.
CREATE TABLE documents (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organisation_id uuid NOT NULL REFERENCES organisations (id),
    type            text NOT NULL,
    file_name       text NOT NULL,
    content_type    text NOT NULL,
    size_bytes      bigint NOT NULL,
    sha256          text NOT NULL,
    storage_key     text NOT NULL UNIQUE,
    uploaded_by     uuid NOT NULL REFERENCES users (id),
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    is_deleted      boolean NOT NULL DEFAULT false
);
CREATE INDEX documents_organisation_id_idx ON documents (organisation_id);
CREATE TRIGGER documents_updated_at BEFORE UPDATE ON documents
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
