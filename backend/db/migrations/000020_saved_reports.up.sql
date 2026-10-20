-- Saved reports: definitions live in the database, generated files in the bucket.
CREATE TABLE saved_reports (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name            text NOT NULL,
    kind            text NOT NULL,
    params          jsonb NOT NULL DEFAULT '{}'::jsonb,
    organisation_id uuid,
    created_by      uuid,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    is_deleted      boolean NOT NULL DEFAULT false
);
CREATE INDEX saved_reports_organisation_id_idx ON saved_reports (organisation_id);

CREATE TABLE report_files (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    saved_id       uuid NOT NULL REFERENCES saved_reports(id),
    format         text NOT NULL,
    bucket_key     text NOT NULL,
    sha256         text NOT NULL,
    size           bigint NOT NULL,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),
    is_deleted     boolean NOT NULL DEFAULT false
);
CREATE INDEX report_files_saved_id_idx ON report_files (saved_id);
CREATE TRIGGER saved_reports_updated_at BEFORE UPDATE ON saved_reports
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER report_files_updated_at BEFORE UPDATE ON report_files
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
