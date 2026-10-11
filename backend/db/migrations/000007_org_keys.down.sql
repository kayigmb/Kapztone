ALTER TABLE organisations
    DROP COLUMN IF EXISTS signing_address,
    DROP COLUMN IF EXISTS signing_key;
