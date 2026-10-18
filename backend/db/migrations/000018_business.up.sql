-- Facilities: a licence with an end date and a document; new facilities wait for approval.
ALTER TABLE facilities
    ADD COLUMN licence_valid_to     date,
    ADD COLUMN licence_file_name    text,
    ADD COLUMN licence_content_type text,
    ADD COLUMN licence_size_bytes   bigint,
    ADD COLUMN licence_sha256       text,
    ADD COLUMN licence_storage_key  text,
    ADD COLUMN approved_by          uuid REFERENCES users (id),
    ADD COLUMN approved_at          timestamptz;

-- Templates: edits wait as a draft until published; every publish is a new version.
ALTER TABLE commodity_templates
    ADD COLUMN version integer NOT NULL DEFAULT 1,
    ADD COLUMN draft   jsonb;

-- The template version a transform or handover was judged by.
ALTER TABLE transforms ADD COLUMN template_version integer;
ALTER TABLE handovers ADD COLUMN template_version integer;

-- Filled in on insert, so every code path that records one gets it.
CREATE FUNCTION set_transform_template_version() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    SELECT version INTO NEW.template_version FROM commodity_templates
    WHERE commodity = NEW.commodity AND NOT is_deleted;
    RETURN NEW;
END $$;
CREATE TRIGGER transforms_template_version BEFORE INSERT ON transforms
    FOR EACH ROW EXECUTE FUNCTION set_transform_template_version();

CREATE FUNCTION set_handover_template_version() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    SELECT t.version INTO NEW.template_version
    FROM passports p JOIN commodity_templates t ON t.id = p.template_id
    WHERE p.id = NEW.passport_id;
    RETURN NEW;
END $$;
CREATE TRIGGER handovers_template_version BEFORE INSERT ON handovers
    FOR EACH ROW EXECUTE FUNCTION set_handover_template_version();

-- Flags: who looks into one, and where it came from (a person, the rules, the public page).
ALTER TABLE flags
    ADD COLUMN assignee_user_id uuid REFERENCES users (id),
    ADD COLUMN source           text NOT NULL DEFAULT 'user';
ALTER TABLE flags ALTER COLUMN source DROP DEFAULT;

-- One row each time a rule refuses an action, for the "blocked" report.
CREATE TABLE rule_events (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organisation_id uuid REFERENCES organisations (id),
    rule            text NOT NULL,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    is_deleted      boolean NOT NULL DEFAULT false
);
CREATE INDEX rule_events_organisation_id_idx ON rule_events (organisation_id);
CREATE INDEX rule_events_created_at_idx ON rule_events (created_at);
CREATE TRIGGER rule_events_updated_at BEFORE UPDATE ON rule_events
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
