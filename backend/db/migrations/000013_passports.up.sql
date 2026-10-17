-- Readable lot codes (ACME-2026-0412): the creating organisation's address in
-- capitals, the year, and a number counted per organisation per year.
CREATE TABLE lot_counters (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organisation_id uuid NOT NULL REFERENCES organisations (id),
    year            integer NOT NULL,
    next            integer NOT NULL,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    is_deleted      boolean NOT NULL DEFAULT false
);
CREATE UNIQUE INDEX lot_counters_org_year_key ON lot_counters (organisation_id, year) WHERE NOT is_deleted;
CREATE TRIGGER lot_counters_updated_at BEFORE UPDATE ON lot_counters
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE passports ADD COLUMN code text;

-- Existing passports get codes in the order they were made. The creator is the
-- origin facility's owner or the transform's organisation, not today's holder.
WITH made AS (
    SELECT p.id, o.id AS org_id, o.slug, extract(year FROM p.created_at AT TIME ZONE 'UTC')::integer AS year,
           row_number() OVER (PARTITION BY o.id, extract(year FROM p.created_at AT TIME ZONE 'UTC')
                              ORDER BY p.created_at, p.id) AS n
    FROM passports p
    LEFT JOIN facilities f ON f.id = p.facility_id
    LEFT JOIN transforms t ON t.id = p.transform_id
    JOIN organisations o ON o.id = COALESCE(f.organisation_id, t.organisation_id)
), coded AS (
    UPDATE passports p SET code = upper(m.slug) || '-' || m.year || '-' ||
        CASE WHEN m.n < 10000 THEN lpad(m.n::text, 4, '0') ELSE m.n::text END
    FROM made m WHERE m.id = p.id
    RETURNING m.org_id, m.year, m.n
)
INSERT INTO lot_counters (organisation_id, year, next)
SELECT org_id, year, max(n) + 1 FROM coded GROUP BY org_id, year;

ALTER TABLE passports ALTER COLUMN code SET NOT NULL;
CREATE UNIQUE INDEX passports_code_key ON passports (code) WHERE NOT is_deleted;

-- The template stage a handover takes the lot into. Older handovers have none.
ALTER TABLE handovers ADD COLUMN stage text NOT NULL DEFAULT '';
ALTER TABLE handovers ALTER COLUMN stage DROP DEFAULT;
