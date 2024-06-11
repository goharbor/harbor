/*
table artifact:
  id            SERIAL PRIMARY KEY NOT NULL,
  type          varchar(255) NOT NULL,
  media_type    varchar(255) NOT NULL,
  manifest_media_type varchar(255) NOT NULL,
  artifact_type varchar(255) NOT NULL,
  project_id    int NOT NULL,
  repository_id int NOT NULL,
  repository_name varchar(255) NOT NULL,
  digest        varchar(255) NOT NULL,
  size          bigint,
  push_time     timestamp default CURRENT_TIMESTAMP,
  pull_time     timestamp,
  extra_attrs   text,
  annotations   jsonb,
  CONSTRAINT unique_artifact UNIQUE (repository_id, digest)
*/

/*
Add new column artifact_type for artifact table to work with oci-spec v1.1.0 list referrer api
*/
ALTER TABLE artifact ADD COLUMN IF NOT EXISTS artifact_type varchar(255);

/*
set value for artifact_type
then set column artifact_type as not null
*/
UPDATE artifact SET artifact_type = media_type WHERE artifact_type IS NULL;

ALTER TABLE artifact ALTER COLUMN artifact_type SET NOT NULL;

CREATE TABLE IF NOT EXISTS sbom_report
(
    id SERIAL PRIMARY KEY NOT NULL,
    uuid VARCHAR(64) UNIQUE NOT NULL,
    artifact_id INT NOT NULL,
    registration_uuid VARCHAR(64) NOT NULL,
    mime_type VARCHAR(256) NOT NULL,
    media_type VARCHAR(256) NOT NULL,
    report JSON,
    UNIQUE(artifact_id, registration_uuid, mime_type, media_type)
);