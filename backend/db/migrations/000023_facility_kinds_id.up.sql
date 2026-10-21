-- Every table carries the four base columns with a uuid id; facility_kinds
-- started life keyed by its natural key instead. Keep the key unique for
-- lookups and duplicate detection, key the table by id like everything else.
ALTER TABLE facility_kinds ADD COLUMN id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE facility_kinds DROP CONSTRAINT facility_kinds_pkey;
ALTER TABLE facility_kinds ADD PRIMARY KEY (id);
ALTER TABLE facility_kinds ADD CONSTRAINT facility_kinds_key_key UNIQUE (key);
