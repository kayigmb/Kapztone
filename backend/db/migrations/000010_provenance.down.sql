DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS flags;
DROP TABLE IF EXISTS attestations;
DROP TABLE IF EXISTS transform_inputs;
-- Passports made by a transform cannot exist without the columns below.
DELETE FROM trace_events WHERE passport_id IN (SELECT id FROM passports WHERE transform_id IS NOT NULL);
DELETE FROM passports WHERE transform_id IS NOT NULL;
ALTER TABLE passports
    DROP COLUMN transform_id,
    ALTER COLUMN facility_id SET NOT NULL,
    ALTER COLUMN period SET NOT NULL;
DROP TABLE IF EXISTS transforms;
ALTER TABLE organisations DROP COLUMN accredited;
