/*
Replication: also tag each copy with the immutable tag of its source revision.
*/
ALTER TABLE replication_policy ADD COLUMN IF NOT EXISTS revision_tag BOOLEAN NOT NULL DEFAULT FALSE;

/*
Replication: optionally rename the copied repository, e.g. a model imported under a name of the user's choice.
*/
ALTER TABLE replication_policy ADD COLUMN IF NOT EXISTS dest_repository VARCHAR(256) NOT NULL DEFAULT '';
