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
