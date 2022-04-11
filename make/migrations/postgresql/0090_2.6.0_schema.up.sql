/* Correct project_metadata.public value, should only be true or false, other invaild value will be rewrite to false */
UPDATE project_metadata SET value='false' WHERE name='public' AND value NOT IN('true', 'false');

CREATE TABLE acceleration_registration
(
    id SERIAL PRIMARY KEY NOT NULL,
    name VARCHAR(128) UNIQUE NOT NULL,
    url VARCHAR(256) NOT NULL,
    access_key VARCHAR(255) NOT NULL,
    access_secret VARCHAR(4096) NOT NULL,
    insecure BOOLEAN NOT NULL DEFAULT FALSE,
    creation_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    credential_type VARCHAR(16) NOT NULL,
    type VARCHAR(128) NOT NULL,
    description VARCHAR(1024) NULL,
    health VARCHAR(16) NOT NULL
);

