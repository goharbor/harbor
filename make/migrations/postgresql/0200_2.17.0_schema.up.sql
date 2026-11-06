/*
Replication: also tag each copy with the immutable tag of its source revision.
*/
ALTER TABLE replication_policy ADD COLUMN IF NOT EXISTS revision_tag BOOLEAN NOT NULL DEFAULT FALSE;

/*
Model sync policies; their runs use the execution/task tables with vendor type MODEL_SYNC.
*/
CREATE TABLE IF NOT EXISTS model_sync_policy (
    id SERIAL PRIMARY KEY NOT NULL,
    name VARCHAR(256) NOT NULL,
    description TEXT,
    creator VARCHAR(256),
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    registry_id INT NOT NULL,
    src_repository VARCHAR(512) NOT NULL,
    src_revision VARCHAR(256),
    file_filter TEXT,
    dest_project_id INT NOT NULL,
    dest_repository VARCHAR(512),
    trigger_type VARCHAR(64) NOT NULL,
    cron VARCHAR(64),
    last_synced_revision VARCHAR(256),
    creation_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_model_sync_policy_name UNIQUE (name)
);
CREATE INDEX IF NOT EXISTS idx_model_sync_policy_registry_id ON model_sync_policy (registry_id);
CREATE INDEX IF NOT EXISTS idx_model_sync_policy_dest_project_id ON model_sync_policy (dest_project_id);
