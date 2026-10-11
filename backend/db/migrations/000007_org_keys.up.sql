-- An organisation's signing key, created when it is verified.
-- signing_address is public; signing_key is the private key, sealed by the application.
ALTER TABLE organisations
    ADD COLUMN signing_address text,
    ADD COLUMN signing_key     text;
