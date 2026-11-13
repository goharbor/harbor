/*
Replication: also tag each copy with the immutable tag of its source revision.
*/
ALTER TABLE replication_policy ADD COLUMN IF NOT EXISTS revision_tag BOOLEAN NOT NULL DEFAULT FALSE;
