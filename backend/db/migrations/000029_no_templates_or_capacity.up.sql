-- Products no longer need a template: a producer names the commodity and unit
-- when issuing. Facilities no longer declare a capacity: the passports record
-- what was produced. Old rows keep their values, because their data hashes
-- cover them; new rows leave them empty.
DROP TRIGGER transforms_template_version ON transforms;
DROP TRIGGER handovers_template_version ON handovers;
DROP FUNCTION set_transform_template_version();
DROP FUNCTION set_handover_template_version();

ALTER TABLE passports DROP CONSTRAINT passports_template_id_fkey;
ALTER TABLE passports ALTER COLUMN template_id DROP NOT NULL;
DROP TABLE commodity_templates;

ALTER TABLE facilities
    ALTER COLUMN capacity DROP NOT NULL,
    ALTER COLUMN capacity_unit DROP NOT NULL,
    ALTER COLUMN capacity_period DROP NOT NULL;
