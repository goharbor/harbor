ALTER TABLE artifact_accessory ADD COLUMN IF NOT EXISTS source varchar(50) DEFAULT 'local';

ALTER TABLE role ADD COLUMN IF NOT EXISTS is_builtin   BOOLEAN      NOT NULL DEFAULT FALSE;
ALTER TABLE role ADD COLUMN IF NOT EXISTS description  TEXT;
ALTER TABLE role ADD COLUMN IF NOT EXISTS modified     BOOLEAN      NOT NULL DEFAULT FALSE;
ALTER TABLE role ADD COLUMN IF NOT EXISTS created_by   VARCHAR(255);
ALTER TABLE role ADD COLUMN IF NOT EXISTS created_at   TIMESTAMP WITH TIME ZONE;
ALTER TABLE role ADD COLUMN IF NOT EXISTS modified_by  VARCHAR(255);
ALTER TABLE role ADD COLUMN IF NOT EXISTS modified_at  TIMESTAMP WITH TIME ZONE;

-- Mark all roles seeded by migrations as built-in (immutable)
UPDATE role SET is_builtin = TRUE
WHERE name IN ('projectAdmin', 'developer', 'guest', 'maintainer', 'limitedGuest');
