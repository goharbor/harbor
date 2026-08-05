ALTER TABLE artifact_accessory ADD COLUMN IF NOT EXISTS source varchar(50) DEFAULT 'local';

/*
Increase the length of the registry access_key column so it can store long-form credentials.

This keeps access_key consistent with access_secret, which is already varchar(4096).

See: https://github.com/goharbor/harbor/issues/23303
*/
ALTER TABLE registry ALTER COLUMN access_key TYPE varchar(4096);

/*
Convert the robot account ID columns to bigint to avoid running out of the int4 range, issue #23091.

The sequence is capped at 9007199254740991 (2^53 - 1, JavaScript's Number.MAX_SAFE_INTEGER) instead of
the bigint maximum, so that every possible robot ID stays exactly representable as a JSON number for
clients that decode it into an IEEE 754 double (browsers, jq, etc.). This still provides ~4 million
times the int4 range, which is far beyond any realistic robot creation rate.
*/
ALTER TABLE robot ALTER COLUMN id TYPE bigint;
ALTER TABLE robot ALTER COLUMN creator_ref TYPE bigint;
ALTER TABLE role_permission ALTER COLUMN role_id TYPE bigint;
ALTER SEQUENCE robot_id_seq AS bigint MAXVALUE 9007199254740991;
