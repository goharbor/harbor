CREATE TABLE registry (
 id SERIAL PRIMARY KEY NOT NULL,
 name varchar(256),
 url varchar(256),
 credential_type varchar(16),
 access_key varchar(128),
 access_secret varchar(1024),
 type varchar(32),
 insecure boolean,
 description varchar(1024),
 health varchar(16),
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP,
 CONSTRAINT unique_registry_name UNIQUE (name)
);