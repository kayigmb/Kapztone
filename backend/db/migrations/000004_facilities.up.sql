-- A place where goods originate or are handled. geom is a point or polygon in WGS 84.
-- capacity is a whole number in capacity_unit, per capacity_period.
CREATE TABLE facilities (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organisation_id uuid NOT NULL REFERENCES organisations (id),
    name            text NOT NULL,
    kind            text NOT NULL,
    geom            geometry(Geometry, 4326) NOT NULL,
    licence_number  text NOT NULL,
    capacity        bigint NOT NULL,
    capacity_unit   text NOT NULL,
    capacity_period text NOT NULL,
    status          text NOT NULL,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    is_deleted      boolean NOT NULL DEFAULT false
);
CREATE INDEX facilities_organisation_id_idx ON facilities (organisation_id);
CREATE INDEX facilities_geom_idx ON facilities USING gist (geom);
CREATE TRIGGER facilities_updated_at BEFORE UPDATE ON facilities
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
