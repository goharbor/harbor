CREATE TABLE dockerfile_optimization
(
    id                          SERIAL PRIMARY KEY NOT NULL,
    repository_name             varchar(256) NOT NULL,
    artifact_digest             varchar(256) NOT NULL,
    dockerfile                  text NOT NULL,
    optimized_dockerfile        text NOT NULL,
    attestation_manifest_digest varchar(256) NOT NULL,
    statement_digest            varchar(256) NOT NULL,
    created_at                  timestamp default CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT unique_dockerfile_optimization UNIQUE (repository_name, artifact_digest)
);

CREATE INDEX IF NOT EXISTS idx_dockerfile_optimization_repo_digest
    ON dockerfile_optimization (repository_name, artifact_digest);
