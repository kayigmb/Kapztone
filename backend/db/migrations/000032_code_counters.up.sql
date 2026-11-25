-- Lot codes become SHORT-YYYY-MM-DD-NNNNNN (PROC-2026-11-25-000004): the first
-- four letters or digits of the organisation's address, the day, and a number.
-- Two organisations can share a prefix, so the number is counted per prefix
-- per year, never per organisation: codes cannot collide.
CREATE TABLE code_counters (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    prefix     text NOT NULL,
    year       integer NOT NULL,
    next       integer NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    is_deleted boolean NOT NULL DEFAULT false
);
CREATE UNIQUE INDEX code_counters_prefix_year_key ON code_counters (prefix, year) WHERE NOT is_deleted;
CREATE TRIGGER code_counters_updated_at BEFORE UPDATE ON code_counters
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Numbering carries on from the organisations' counts so far.
INSERT INTO code_counters (prefix, year, next)
SELECT left(upper(regexp_replace(o.slug, '[^a-zA-Z0-9]', '', 'g')), 4), c.year, max(c.next)
FROM lot_counters c JOIN organisations o ON o.id = c.organisation_id
WHERE NOT c.is_deleted
GROUP BY 1, 2;
