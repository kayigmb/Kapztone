-- Changes waiting to be written to the chain. A row is inserted in the same
-- transaction as the change it describes, and processed in seq order.
CREATE TABLE chain_outbox (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    seq        bigint GENERATED ALWAYS AS IDENTITY UNIQUE,
    kind       text NOT NULL,
    subject_id uuid NOT NULL,
    payload    jsonb NOT NULL,
    status     text NOT NULL,
    attempts   int NOT NULL DEFAULT 0,
    last_error text,
    tx_hash    text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    is_deleted boolean NOT NULL DEFAULT false
);
CREATE INDEX chain_outbox_status_seq_idx ON chain_outbox (status, seq);
CREATE TRIGGER chain_outbox_updated_at BEFORE UPDATE ON chain_outbox
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Hash of the facility record as it was registered on-chain.
ALTER TABLE facilities ADD COLUMN data_hash text;
