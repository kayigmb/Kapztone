ALTER TABLE facility_kinds DROP CONSTRAINT facility_kinds_key_key;
ALTER TABLE facility_kinds DROP CONSTRAINT facility_kinds_pkey;
ALTER TABLE facility_kinds DROP COLUMN id;
ALTER TABLE facility_kinds ADD PRIMARY KEY (key);
