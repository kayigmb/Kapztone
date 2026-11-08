-- One-way in practice: templates and capacities are gone. This only restores
-- the shape, so earlier migrations can run down.
CREATE TABLE commodity_templates (
    id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    commodity            text NOT NULL,
    name                 text NOT NULL,
    unit                 text NOT NULL,
    origin_facility_kind text NOT NULL,
    capacity_rule        jsonb NOT NULL,
    required_evidence    text[] NOT NULL DEFAULT '{}',
    version              integer NOT NULL DEFAULT 1,
    draft                jsonb,
    is_deleted           boolean NOT NULL DEFAULT false,
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now()
);
CREATE FUNCTION set_transform_template_version() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN RETURN NEW; END $$;
CREATE TRIGGER transforms_template_version BEFORE INSERT ON transforms
    FOR EACH ROW EXECUTE FUNCTION set_transform_template_version();
CREATE FUNCTION set_handover_template_version() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN RETURN NEW; END $$;
CREATE TRIGGER handovers_template_version BEFORE INSERT ON handovers
    FOR EACH ROW EXECUTE FUNCTION set_handover_template_version();
