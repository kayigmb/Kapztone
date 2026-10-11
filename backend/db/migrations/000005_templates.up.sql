-- Platform-wide rules for one commodity: its unit, stages, allowed events,
-- yields and capacity rule. Not owned by an organisation.
CREATE TABLE commodity_templates (
    id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    commodity            text NOT NULL,
    name                 text NOT NULL,
    unit                 text NOT NULL,
    origin_facility_kind text NOT NULL,
    capacity_rule        jsonb NOT NULL,
    stages               jsonb NOT NULL,
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),
    is_deleted           boolean NOT NULL DEFAULT false
);
CREATE UNIQUE INDEX commodity_templates_commodity_key ON commodity_templates (commodity) WHERE NOT is_deleted;
CREATE TRIGGER commodity_templates_updated_at BEFORE UPDATE ON commodity_templates
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
