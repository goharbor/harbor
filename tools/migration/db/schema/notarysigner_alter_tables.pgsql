\c notarysigner;

ALTER TABLE private_keys OWNER TO signer;
ALTER SEQUENCE private_keys_id_seq OWNER TO signer;
ALTER TABLE schema_migrations OWNER TO signer;