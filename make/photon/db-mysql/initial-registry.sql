CREATE DATABASE if NOT EXISTS `registry` default character set utf8mb4 collate utf8mb4_unicode_ci;

USE `registry`;
CREATE TABLE schema_migrations(version bigint not null primary key, dirty boolean not null);