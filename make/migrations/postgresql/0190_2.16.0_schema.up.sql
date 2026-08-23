ALTER TABLE artifact_accessory ADD COLUMN IF NOT EXISTS source varchar(50) DEFAULT 'local';

/*
Increase the length of the registry access_key column so it can store long-form credentials.

This keeps access_key consistent with access_secret, which is already varchar(4096).

See: https://github.com/goharbor/harbor/issues/23303
*/
ALTER TABLE registry ALTER COLUMN access_key TYPE varchar(4096);

/*
Convert the robot account ID columns to bigint to avoid running out of the int4 range, issue #23091.
*/
ALTER TABLE robot ALTER COLUMN id TYPE bigint;
ALTER TABLE robot ALTER COLUMN creator_ref TYPE bigint;
ALTER TABLE role_permission ALTER COLUMN role_id TYPE bigint;
ALTER SEQUENCE robot_id_seq AS bigint MAXVALUE 9007199254740991;
