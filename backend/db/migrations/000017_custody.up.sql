-- Handing over part of a lot: the part is first split into its own passport,
-- and that child is handed over. split_from is the passport it was cut from.
ALTER TABLE handovers ADD COLUMN split_from uuid REFERENCES passports (id);

-- Where the sender handed the goods over, and what the receiver weighed and where.
ALTER TABLE handovers ADD COLUMN lat double precision;
ALTER TABLE handovers ADD COLUMN lon double precision;
ALTER TABLE handovers ADD COLUMN received_quantity bigint;
ALTER TABLE handovers ADD COLUMN receiver_lat double precision;
ALTER TABLE handovers ADD COLUMN receiver_lon double precision;

-- Why a receiver refused the goods (from a declared list); reject_reason keeps the note.
ALTER TABLE handovers ADD COLUMN reject_reasons text[] NOT NULL DEFAULT '{}';

-- Photos taken at a handover, by the sender when sending or the receiver when accepting.
CREATE TABLE handover_photos (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    handover_id  uuid NOT NULL REFERENCES handovers (id),
    side         text NOT NULL,
    file_name    text NOT NULL,
    content_type text NOT NULL,
    size_bytes   bigint NOT NULL,
    sha256       text NOT NULL,
    storage_key  text NOT NULL,
    uploaded_by  uuid NOT NULL REFERENCES users (id),
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    is_deleted   boolean NOT NULL DEFAULT false
);
CREATE INDEX handover_photos_handover_id_idx ON handover_photos (handover_id);
CREATE TRIGGER handover_photos_updated_at BEFORE UPDATE ON handover_photos
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
