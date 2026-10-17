ALTER TABLE organisations ADD COLUMN accredited boolean NOT NULL DEFAULT false;
UPDATE organisations SET accredited = true
    WHERE id IN (SELECT organisation_id FROM accreditations WHERE NOT is_deleted);
DROP TABLE IF EXISTS accreditations;
