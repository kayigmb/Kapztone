-- A certificate whose IPFS copy failed is tried again later, backing off.
ALTER TABLE attestations
    ADD COLUMN ipfs_attempts   int NOT NULL DEFAULT 0,
    ADD COLUMN ipfs_next_at    timestamptz,
    ADD COLUMN ipfs_last_error text;

CREATE INDEX attestations_ipfs_due_idx ON attestations (ipfs_next_at) WHERE ipfs_cid IS NULL AND NOT is_deleted;
