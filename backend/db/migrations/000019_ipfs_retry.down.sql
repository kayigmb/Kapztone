DROP INDEX IF EXISTS attestations_ipfs_due_idx;
ALTER TABLE attestations DROP COLUMN ipfs_last_error, DROP COLUMN ipfs_next_at, DROP COLUMN ipfs_attempts;
