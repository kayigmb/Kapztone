-- Facility kinds the platform curates: what a facility can be registered as.
-- Templates name an origin kind, so a kind in use cannot be deleted, only
-- deactivated (new facilities and templates then stop offering it).
CREATE TABLE facility_kinds (
    key         text PRIMARY KEY,
    label       text NOT NULL,
    sub         text NOT NULL DEFAULT '',
    active      boolean NOT NULL DEFAULT true,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    is_deleted  boolean NOT NULL DEFAULT false
);
CREATE TRIGGER facility_kinds_updated_at BEFORE UPDATE ON facility_kinds
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

INSERT INTO facility_kinds (key, label, sub) VALUES
    ('farm_plot', 'Farm plot', 'Land growing a crop'),
    ('mine_site', 'Mine site', 'Extraction site'),
    ('washing_station', 'Washing station', 'Coffee or tea processing'),
    ('processing_plant', 'Processing plant', 'Factory, mill or smelter'),
    ('warehouse', 'Warehouse', 'Storage and depot');
