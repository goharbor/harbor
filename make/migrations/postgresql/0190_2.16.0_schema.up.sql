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

CREATE INDEX IF NOT EXISTS idx_sbom_report_sbom_digest
  ON sbom_report (mime_type, ((report::jsonb ->> 'sbom_digest')));

/*
Custom project roles: schema additions on the existing `role` table.

Built-in role permissions are NOT stored in the database — they are resolved from
the compile-time rolePoliciesMap in common/rbac/project, so authorization for
built-in roles incurs no DB lookup. Only custom roles persist their permissions
(in permission_policy/role_permission).
*/
ALTER TABLE role ADD COLUMN IF NOT EXISTS is_builtin   BOOLEAN      NOT NULL DEFAULT FALSE;
ALTER TABLE role ADD COLUMN IF NOT EXISTS description  TEXT;
ALTER TABLE role ADD COLUMN IF NOT EXISTS modified     BOOLEAN      NOT NULL DEFAULT FALSE;
ALTER TABLE role ADD COLUMN IF NOT EXISTS created_by   VARCHAR(255);
ALTER TABLE role ADD COLUMN IF NOT EXISTS created_at   TIMESTAMP WITH TIME ZONE;
ALTER TABLE role ADD COLUMN IF NOT EXISTS modified_by  VARCHAR(255);
ALTER TABLE role ADD COLUMN IF NOT EXISTS modified_at  TIMESTAMP WITH TIME ZONE;

-- Widen the role name to match the API/UI contract (was varchar(20)) and enforce
-- name uniqueness so custom roles cannot collide.
ALTER TABLE role ALTER COLUMN name TYPE varchar(255);
CREATE UNIQUE INDEX IF NOT EXISTS uq_role_name ON role (name);

-- Mark all roles seeded by migrations as built-in (immutable).
UPDATE role SET is_builtin = TRUE
WHERE name IN ('projectAdmin', 'developer', 'guest', 'maintainer', 'limitedGuest');
