DROP TABLE IF EXISTS documents;
ALTER TABLE organisations
    DROP COLUMN IF EXISTS reviewed_by,
    DROP COLUMN IF EXISTS reviewed_at,
    DROP COLUMN IF EXISTS review_reason;
