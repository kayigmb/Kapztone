-- M4 small gaps: flag code/severity, cert number/notes, handover seal/checklist.
ALTER TABLE flags
    ADD COLUMN code text,
    ADD COLUMN severity text NOT NULL DEFAULT 'medium',
    ADD COLUMN request_info text;
CREATE UNIQUE INDEX flags_code_idx ON flags (code) WHERE code IS NOT NULL;

ALTER TABLE attestations
    ADD COLUMN cert_number text,
    ADD COLUMN notes text,
    ADD COLUMN max_valid_days int;
CREATE UNIQUE INDEX attestations_cert_number_idx ON attestations (cert_number) WHERE cert_number IS NOT NULL;

ALTER TABLE handovers
    ADD COLUMN seal_number text,
    ADD COLUMN checklist jsonb NOT NULL DEFAULT '{}'::jsonb;
