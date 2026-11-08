-- Where a certificate was issued: one of the issuer's own facilities (an
-- office or lab), optional.
ALTER TABLE attestations ADD COLUMN issued_at_facility_id uuid REFERENCES facilities (id);
