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
-- name uniqueness so custom roles cannot collide. The uniqueness is
-- case-insensitive (lower(name)) so "Maintainer" and "maintainer" cannot coexist
-- and be mistaken for one another.
ALTER TABLE role ALTER COLUMN name TYPE varchar(255);
DROP INDEX IF EXISTS uq_role_name;
CREATE UNIQUE INDEX IF NOT EXISTS uq_role_name ON role (lower(name));

-- Mark all roles seeded by migrations as built-in (immutable).
UPDATE role SET is_builtin = TRUE
WHERE name IN ('projectAdmin', 'developer', 'guest', 'maintainer', 'limitedGuest');

-- Referential integrity between a project member and its role. ON DELETE RESTRICT
-- makes the database reject deleting a role that is still assigned to a member,
-- closing the count-then-delete race in the role Delete controller (a concurrent
-- member assignment can no longer leave a dangling project_member.role).
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_project_member_role'
    ) THEN
        ALTER TABLE project_member
            ADD CONSTRAINT fk_project_member_role
            FOREIGN KEY (role) REFERENCES role (role_id) ON DELETE RESTRICT;
    END IF;
END $$;

/*
Every core replica holds the role policy in memory and reads the database when
that policy changes, not when a user asks a permission question. Harbor runs
several cores against one database, so a replica that misses a change and never
finds out would keep authorizing against a policy that no longer exists.

Two things guard against that, and both are driven from here.

The trigger sends NOTIFY from inside the writing transaction, so it cannot
survive a rollback and it fires for writes that never went through Harbor, such
as a migration or a support script. Its payload carries no rules, only the new
version and the replica that caused it, so a replica goes back to the table
rather than applying something it was handed, and skips its own writes.

The same statement bumps policy_version. A replica records the version it last
loaded and compares it against this counter, so a notification that never
arrives, a dropped connection or a failed reload all converge anyway.
*/
CREATE TABLE IF NOT EXISTS policy_version (
  only_row    boolean   PRIMARY KEY DEFAULT TRUE,
  version     bigint    NOT NULL DEFAULT 1,
  update_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT policy_version_single_row CHECK (only_row)
);
INSERT INTO policy_version (only_row) VALUES (TRUE) ON CONFLICT DO NOTHING;

CREATE OR REPLACE FUNCTION harbor_policy_bump() RETURNS void AS $$
DECLARE
  v bigint;
BEGIN
  UPDATE policy_version
     SET version = version + 1, update_time = CURRENT_TIMESTAMP
   WHERE only_row
  RETURNING version INTO v;

  PERFORM pg_notify(
    'harbor_policy',
    v::text || ':' || coalesce(current_setting('harbor.origin', TRUE), '?'));
END;
$$ LANGUAGE plpgsql;

/*
Both tables are shared with robot accounts, which write one row per grant every
time a robot is created or edited. A robot has nothing to do with what a
project role grants, so the triggers below look at the rows the statement
actually touched and stay quiet unless a project-role link is among them.
Without that test, creating a robot would bump the version and make every core
in the fleet reload a policy that did not change.

Transition tables are what make that test possible from a statement-level
trigger, and a trigger carrying one may name a single event, so there is one
trigger per event rather than one for all three.
*/
CREATE OR REPLACE FUNCTION harbor_role_permission_notify() RETURNS trigger AS $$
BEGIN
  IF EXISTS (SELECT 1 FROM changed WHERE role_type = 'project-role') THEN
    PERFORM harbor_policy_bump();
  END IF;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- A policy row matters to the store only while a project role links to it. A
-- robot's own policy rows are invisible here, and a policy inserted before
-- anything links to it is picked up by the link's own trigger.
CREATE OR REPLACE FUNCTION harbor_permission_policy_notify() RETURNS trigger AS $$
BEGIN
  IF EXISTS (
    SELECT 1
      FROM role_permission rp
      JOIN changed c ON c.id = rp.permission_policy_id
     WHERE rp.role_type = 'project-role'
  ) THEN
    PERFORM harbor_policy_bump();
  END IF;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- AFTER STATEMENT, not AFTER ROW: replacing a role's permissions is one delete
-- and one insert, and the fleet does not need to rebuild once per row.
DROP TRIGGER IF EXISTS role_permission_notify ON role_permission;
DROP TRIGGER IF EXISTS role_permission_notify_ins ON role_permission;
DROP TRIGGER IF EXISTS role_permission_notify_upd ON role_permission;
DROP TRIGGER IF EXISTS role_permission_notify_del ON role_permission;
CREATE TRIGGER role_permission_notify_ins
  AFTER INSERT ON role_permission
  REFERENCING NEW TABLE AS changed
  FOR EACH STATEMENT EXECUTE FUNCTION harbor_role_permission_notify();
CREATE TRIGGER role_permission_notify_upd
  AFTER UPDATE ON role_permission
  REFERENCING NEW TABLE AS changed
  FOR EACH STATEMENT EXECUTE FUNCTION harbor_role_permission_notify();
CREATE TRIGGER role_permission_notify_del
  AFTER DELETE ON role_permission
  REFERENCING OLD TABLE AS changed
  FOR EACH STATEMENT EXECUTE FUNCTION harbor_role_permission_notify();

DROP TRIGGER IF EXISTS permission_policy_notify ON permission_policy;
DROP TRIGGER IF EXISTS permission_policy_notify_ins ON permission_policy;
DROP TRIGGER IF EXISTS permission_policy_notify_upd ON permission_policy;
DROP TRIGGER IF EXISTS permission_policy_notify_del ON permission_policy;
CREATE TRIGGER permission_policy_notify_ins
  AFTER INSERT ON permission_policy
  REFERENCING NEW TABLE AS changed
  FOR EACH STATEMENT EXECUTE FUNCTION harbor_permission_policy_notify();
CREATE TRIGGER permission_policy_notify_upd
  AFTER UPDATE ON permission_policy
  REFERENCING NEW TABLE AS changed
  FOR EACH STATEMENT EXECUTE FUNCTION harbor_permission_policy_notify();
CREATE TRIGGER permission_policy_notify_del
  AFTER DELETE ON permission_policy
  REFERENCING OLD TABLE AS changed
  FOR EACH STATEMENT EXECUTE FUNCTION harbor_permission_policy_notify();
