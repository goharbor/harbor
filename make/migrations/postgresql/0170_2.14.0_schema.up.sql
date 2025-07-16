ALTER TABLE role_permission ALTER COLUMN id TYPE BIGINT;
ALTER SEQUENCE role_permission_id_seq AS BIGINT;

ALTER TABLE permission_policy ALTER COLUMN id TYPE BIGINT;
ALTER SEQUENCE permission_policy_id_seq AS BIGINT;

ALTER TABLE role_permission ALTER COLUMN permission_policy_id TYPE BIGINT;

ALTER TABLE vulnerability_record ADD COLUMN IF NOT EXISTS status text;