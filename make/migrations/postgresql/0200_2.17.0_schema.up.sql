CREATE TABLE IF NOT EXISTS model_import_policy (
    id SERIAL PRIMARY KEY NOT NULL,
    name varchar(255) NOT NULL,
    description varchar(1024),
    project_id int NOT NULL,
    target_repository varchar(255) NOT NULL,
    target_tag varchar(255) NOT NULL DEFAULT 'latest',
    hf_repo_id varchar(255) NOT NULL,
    hf_revision varchar(255) NOT NULL DEFAULT 'main',
    hf_subpath varchar(1024) DEFAULT '',
    hf_token varchar(4096) DEFAULT '',
    last_synced_commit varchar(255) DEFAULT '',
    last_artifact_digest varchar(255) DEFAULT '',
    creator varchar(255) DEFAULT '',
    enabled boolean NOT NULL DEFAULT true,
    trigger text NOT NULL,
    extra_attrs text,
    creation_time timestamp default CURRENT_TIMESTAMP,
    update_time timestamp default CURRENT_TIMESTAMP,
    CONSTRAINT unique_model_import_policy_name UNIQUE (project_id, name),
    CONSTRAINT fk_model_import_policy_project FOREIGN KEY (project_id) REFERENCES project (project_id)
);

CREATE INDEX IF NOT EXISTS idx_model_import_policy_project_id ON model_import_policy (project_id);
CREATE INDEX IF NOT EXISTS idx_model_import_policy_hf_repo_id ON model_import_policy (hf_repo_id);
