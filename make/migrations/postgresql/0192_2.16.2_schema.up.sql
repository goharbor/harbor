/*
Pluggable optimizer adapter framework.

optimizer_registration mirrors scanner_registration (0015_1.10.0_schema.up.sql):
optimizer adapters (e.g. the sleeko Dockerfile optimizer) are registered as
external services that Harbor talks to over a versioned REST contract.
*/
CREATE TABLE optimizer_registration
(
    id SERIAL PRIMARY KEY NOT NULL,
    uuid VARCHAR(64) UNIQUE NOT NULL,
    url VARCHAR(512) UNIQUE NOT NULL,
    name VARCHAR(128) UNIQUE NOT NULL,
    description VARCHAR(1024) NULL,
    auth VARCHAR(16) NOT NULL,
    access_cred VARCHAR(512) NULL,
    disabled BOOLEAN NOT NULL DEFAULT FALSE,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    use_internal_addr BOOLEAN NOT NULL DEFAULT FALSE,
    immutable BOOLEAN NOT NULL DEFAULT FALSE,
    skip_cert_verify BOOLEAN NOT NULL DEFAULT FALSE,
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

/*
Optimization results become asynchronous: rows are created in Pending state when
the job is launched and filled in when the adapter reports back. Existing rows
were produced by the old synchronous flow, so they backfill as Success.
*/
ALTER TABLE dockerfile_optimization
    ADD COLUMN status varchar(16) NOT NULL DEFAULT 'Success',
    ADD COLUMN error text NOT NULL DEFAULT '',
    ADD COLUMN registration_uuid varchar(64) NOT NULL DEFAULT '',
    ADD COLUMN execution_id bigint NOT NULL DEFAULT 0,
    ADD COLUMN update_time timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL;
