CREATE DATABASE registry ENCODING 'UTF8';
\c registry;

CREATE TABLE schema_migrations(version bigint not null primary key, dirty boolean not null);