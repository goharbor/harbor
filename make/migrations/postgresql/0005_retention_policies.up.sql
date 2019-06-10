CREATE TABLE IF NOT EXISTS retention_policy (
    id SERIAL PRIMARY KEY NOT NULL,
    name VARCHAR(255) NOT NULL,
    enabled BOOLEAN NOT NULL,

    scope INT NOT NULL,
    fall_through_action INT NOT NULL,

    project_id INTEGER REFERENCES project(project_id) ON DELETE CASCADE,
    repository_id INTEGER REFERENCES repository(repository_id) ON DELETE CASCADE,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS retention_policy_unique ON retention_policy (
    COALESCE(project_id, 0),
    COALESCE(repository_id, 0)
);

CREATE TABLE IF NOT EXISTS retention_filter_metadata (
    id SERIAL PRIMARY KEY NOT NULL,
    type VARCHAR(128) NOT NULL,

    options JSON,

    policy INTEGER REFERENCES retention_policy(id) ON DELETE CASCADE
);