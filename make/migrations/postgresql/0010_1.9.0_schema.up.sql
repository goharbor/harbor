/* add table for CVE whitelist */
CREATE TABLE cve_whitelist (
    id SERIAL PRIMARY KEY NOT NULL,
    project_id int,
    creation_time timestamp default CURRENT_TIMESTAMP,
    update_time timestamp default CURRENT_TIMESTAMP,
    expires_at bigint,
    items text NOT NULL,
    UNIQUE (project_id)
);

CREATE TABLE blob (
 id SERIAL PRIMARY KEY NOT NULL,
 /*
    digest of config, layer, manifest
 */
 digest varchar(255) NOT NULL,
 content_type varchar(255) NOT NULL,
 size int NOT NULL,
 creation_time timestamp default CURRENT_TIMESTAMP,
 UNIQUE (digest)
);

CREATE TABLE artifact (
 id SERIAL PRIMARY KEY NOT NULL,
 project_id int NOT NULL,
 repo varchar(255) NOT NULL,
 tag varchar(255) NOT NULL,
 /*
    digest of manifest
 */
 digest varchar(255) NOT NULL,
 /*
    kind of artifact, image, chart, etc..
 */
 kind varchar(255) NOT NULL,
 creation_time timestamp default CURRENT_TIMESTAMP,
 pull_time timestamp,
 push_time timestamp,
 CONSTRAINT unique_artifact UNIQUE (project_id, repo, tag)
);

/* add the table for relation of artifact and blob */
CREATE TABLE artifact_blob (
 id SERIAL PRIMARY KEY NOT NULL,
 digest_af varchar(255) NOT NULL,
 digest_blob varchar(255) NOT NULL,
 creation_time timestamp default CURRENT_TIMESTAMP,
 CONSTRAINT unique_artifact_blob UNIQUE (digest_af, digest_blob)
);

/* add quota table */
CREATE TABLE quota (
 id SERIAL PRIMARY KEY NOT NULL,
 reference VARCHAR(255) NOT NULL,
 reference_id VARCHAR(255) NOT NULL,
 hard JSONB NOT NULL,
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP,
 UNIQUE(reference, reference_id)
);

/* add quota usage table */
CREATE TABLE quota_usage (
 id SERIAL PRIMARY KEY NOT NULL,
 reference VARCHAR(255) NOT NULL,
 reference_id VARCHAR(255) NOT NULL,
 used JSONB NOT NULL,
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP,
 UNIQUE(reference, reference_id)
);
