DROP TABLE IF EXISTS password_resets;
DROP TABLE IF EXISTS password_history;
ALTER TABLE users DROP COLUMN IF EXISTS token_version;
