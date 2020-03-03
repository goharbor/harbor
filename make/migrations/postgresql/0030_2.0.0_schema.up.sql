/*
table artifact:
  id            SERIAL PRIMARY KEY NOT NULL,
  type          varchar(255) NOT NULL,
  media_type    varchar(255) NOT NULL,
  manifest_media_type varchar(255) NOT NULL,
  project_id    int NOT NULL,
  repository_id int NOT NULL,
  repository_name varchar(255) NOT NULL,
  digest        varchar(255) NOT NULL,
  size          bigint,
  push_time     timestamp default CURRENT_TIMESTAMP,
  pull_time     timestamp,
  extra_attrs   text,
  annotations   jsonb,
  CONSTRAINT unique_artifact_2 UNIQUE (repository_id, digest)
*/

ALTER TABLE admin_job ADD COLUMN job_parameters varchar(255) Default '';

/*record the data version to decide whether the data migration should be skipped*/
ALTER TABLE schema_migrations ADD COLUMN data_version int;

ALTER TABLE artifact ADD COLUMN repository_id int;
ALTER TABLE artifact ADD COLUMN media_type varchar(255);
ALTER TABLE artifact ADD COLUMN manifest_media_type varchar(255);
ALTER TABLE artifact ADD COLUMN size bigint;
ALTER TABLE artifact ADD COLUMN extra_attrs text;
ALTER TABLE artifact ADD COLUMN annotations jsonb;
ALTER TABLE artifact RENAME COLUMN kind TO type;
ALTER TABLE artifact DROP COLUMN creation_time;

/*set the media type*/
UPDATE artifact AS art
    SET type='IMAGE', repository_id=repo.repository_id,
    manifest_media_type=blob.content_type,
    media_type=(
    CASE
        /*v2 manifest*/
        WHEN blob.content_type='application/vnd.docker.distribution.manifest.v2+json' THEN
            'application/vnd.docker.container.image.v1+json'
        /*manifest list*/
        WHEN blob.content_type='application/vnd.docker.distribution.manifest.list.v2+json' THEN
            'application/vnd.docker.distribution.manifest.list.v2+json'
        /*v1 manifest*/
        ELSE
            'application/vnd.docker.distribution.manifest.v1+prettyjws'
    END
    )
    FROM repository AS repo, blob AS blob
    WHERE art.repo=repo.name AND art.digest=blob.digest;
ALTER TABLE artifact ALTER COLUMN repository_id SET NOT NULL;
ALTER TABLE artifact ALTER COLUMN media_type SET NOT NULL;
ALTER TABLE artifact ALTER COLUMN manifest_media_type SET NOT NULL;
ALTER TABLE artifact RENAME COLUMN repo TO repository_name;

CREATE TABLE tag
(
  id            SERIAL PRIMARY KEY NOT NULL,
  repository_id int NOT NULL,
  artifact_id   int NOT NULL,
  name          varchar(255) NOT NULL,
  push_time     timestamp default CURRENT_TIMESTAMP,
  pull_time     timestamp,
  FOREIGN KEY (artifact_id) REFERENCES artifact(id),
  CONSTRAINT unique_tag UNIQUE (repository_id, name)
);

/*move the tag in the table artifact into table tag*/
INSERT INTO tag (artifact_id, repository_id, name, push_time, pull_time)
SELECT ordered_art.id, art.repository_id, art.tag, art.push_time, art.pull_time
FROM artifact AS art
JOIN (
    /*the tag references the first artifact that with the same digest*/
    SELECT id, repository_name, digest, row_number() OVER (PARTITION BY repository_name, digest ORDER BY id) AS seq FROM artifact
) AS ordered_art ON art.repository_name=ordered_art.repository_name AND art.digest=ordered_art.digest
WHERE ordered_art.seq=1;

ALTER TABLE artifact DROP COLUMN tag;

/*remove the duplicate artifact rows*/
DELETE FROM artifact
WHERE id NOT IN (
    SELECT artifact_id
    FROM tag
);

ALTER TABLE artifact ADD CONSTRAINT unique_artifact UNIQUE (repository_id, digest);

/*set artifact size*/
UPDATE artifact
SET size=s.size
FROM (
    SELECT art.digest, sum(blob.size) AS size
        FROM artifact AS art, artifact_blob AS ref, blob AS blob
        WHERE art.digest=ref.digest_af AND ref.digest_blob=blob.digest
        GROUP BY art.digest
) AS s
WHERE artifact.digest=s.digest;


/* artifact_reference records the child artifact referenced by parent artifact */
CREATE TABLE artifact_reference
(
  id          SERIAL PRIMARY KEY NOT NULL,
  parent_id   int NOT NULL,
  child_id    int NOT NULL,
  platform    varchar(255),
  FOREIGN KEY (parent_id) REFERENCES artifact(id),
  FOREIGN KEY (child_id) REFERENCES artifact(id),
  CONSTRAINT  unique_reference UNIQUE (parent_id, child_id)
);

/* artifact_trash records deleted artifact */
CREATE TABLE artifact_trash
(
  id                  SERIAL PRIMARY KEY NOT NULL,
  media_type          varchar(255) NOT NULL,
  manifest_media_type varchar(255) NOT NULL,
  repository_name     varchar(255) NOT NULL,
  digest              varchar(255) NOT NULL,
  creation_time       timestamp default CURRENT_TIMESTAMP,
  CONSTRAINT      unique_artifact_trash UNIQUE (repository_name, digest)
);

/* label_reference records the labels added to the artifact */
CREATE TABLE label_reference (
 id SERIAL PRIMARY KEY NOT NULL,
 label_id int NOT NULL,
 artifact_id int NOT NULL,
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP,
 FOREIGN KEY (label_id) REFERENCES harbor_label(id),
 FOREIGN KEY (artifact_id) REFERENCES artifact(id),
 CONSTRAINT unique_label_reference UNIQUE (label_id,artifact_id)
);

/*move the labels added to tag to artifact*/
INSERT INTO label_reference (label_id, artifact_id, creation_time, update_time)
(
SELECT label.label_id, repo_tag.artifact_id, label.creation_time, label.update_time
    FROM harbor_resource_label AS label
    JOIN (
        SELECT tag.artifact_id, CONCAT(repository.name, ':', tag.name) as name
            FROM tag
            JOIN repository
            ON tag.repository_id = repository.repository_id
    ) AS repo_tag
    ON repo_tag.name = label.resource_name AND label.resource_type = 'i'
) ON CONFLICT DO NOTHING;

/*remove the records for images in table 'harbor_resource_label'*/
DELETE FROM harbor_resource_label WHERE resource_type = 'i';


/* TODO remove this table after clean up code that related with the old artifact model */
CREATE TABLE artifact_2
(
  id            SERIAL PRIMARY KEY NOT NULL,
  project_id    int                NOT NULL,
  repo          varchar(255)       NOT NULL,
  tag           varchar(255)       NOT NULL,
  /*
     digest of manifest
  */
  digest        varchar(255)       NOT NULL,
  /*
     kind of artifact, image, chart, etc..
  */
  kind          varchar(255)       NOT NULL,
  creation_time timestamp default CURRENT_TIMESTAMP,
  pull_time     timestamp,
  push_time     timestamp,
  CONSTRAINT unique_artifact_2 UNIQUE (project_id, repo, tag)
);

/*remove the constraint for project_id in table 'notification_policy'*/
ALTER TABLE notification_policy DROP CONSTRAINT unique_project_id;