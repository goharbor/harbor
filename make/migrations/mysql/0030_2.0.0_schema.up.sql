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
  CONSTRAINT unique_artifact UNIQUE (repository_id, digest)
*/

ALTER TABLE admin_job ADD COLUMN job_parameters varchar(255) Default '';

/*record the data version to decide whether the data migration should be skipped*/
ALTER TABLE schema_migrations ADD COLUMN data_version int;

ALTER TABLE artifact ADD COLUMN repository_id int;
ALTER TABLE artifact ADD COLUMN media_type varchar(255);
ALTER TABLE artifact ADD COLUMN manifest_media_type varchar(255);
ALTER TABLE artifact ADD COLUMN size bigint;
ALTER TABLE artifact ADD COLUMN extra_attrs text;
ALTER TABLE artifact ADD COLUMN annotations json;
ALTER TABLE artifact CHANGE COLUMN kind type varchar(255) NOT NULL;
ALTER TABLE artifact DROP COLUMN creation_time;

/*set the media type*/
UPDATE artifact AS art, repository AS repo, `blob`
    SET type='IMAGE', art.repository_id=repo.repository_id,
    manifest_media_type=`blob`.content_type,
    media_type=(
    CASE
        /*v2 manifest*/
        WHEN `blob`.content_type='application/vnd.docker.distribution.manifest.v2+json' THEN
            'application/vnd.docker.container.image.v1+json'
        /*manifest list*/
        WHEN `blob`.content_type='application/vnd.docker.distribution.manifest.list.v2+json' THEN
            'application/vnd.docker.distribution.manifest.list.v2+json'
        /*v1 manifest*/
        ELSE
            'application/vnd.docker.distribution.manifest.v1+prettyjws'
    END
    )
    WHERE art.repo=repo.name AND art.digest=`blob`.digest;
/*
It's a workaround for issue https://github.com/goharbor/harbor/issues/11754

The phenomenon is the repository data is gone, but artifacts belong to the repository are still there.
To set the repository_id to a negative, and cannot duplicate.
*/
UPDATE artifact SET repository_id = 0-artifact.id, type='IMAGE', media_type='UNKNOWN', manifest_media_type='UNKNOWN' WHERE repository_id IS NULL;

ALTER TABLE artifact MODIFY COLUMN repository_id INT NOT NULL;
ALTER TABLE artifact MODIFY COLUMN media_type varchar(255) NOT NULL;
ALTER TABLE artifact MODIFY COLUMN manifest_media_type varchar(255) NOT NULL;
ALTER TABLE artifact CHANGE COLUMN repo repository_name varchar(255) NOT NULL;

CREATE TABLE tag
(
  id            SERIAL PRIMARY KEY NOT NULL,
  repository_id int NOT NULL,
  artifact_id   bigint unsigned NOT NULL,
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
		SELECT
			id,
			repository_name,
			digest,
			@idx := IF(@gid = repository_name and @did = digest,@idx + 1, 1) as seq,
			@gid := repository_name,
			@did := digest
		FROM
			artifact, (SELECT @idx := 0, @gid := NULL, @did := NULL) t
			order by repository_name, digest, id
) AS ordered_art ON art.repository_name=ordered_art.repository_name AND art.digest=ordered_art.digest
WHERE ordered_art.seq=1;

ALTER TABLE artifact DROP COLUMN tag;

/*remove the duplicate artifact rows*/
DELETE FROM artifact
WHERE id NOT IN (
    SELECT artifact_id
    FROM tag
);

SET sql_mode = '';
ALTER TABLE artifact DROP INDEX unique_artifact;
ALTER TABLE artifact ADD CONSTRAINT unique_artifact UNIQUE (repository_id, digest);

/*set artifact size*/
UPDATE artifact ,(
    SELECT art.digest, sum(blob.size) AS size
        FROM artifact AS art, artifact_blob AS ref, `blob`
        WHERE art.digest=ref.digest_af AND ref.digest_blob=`blob`.digest
        GROUP BY art.digest
) AS s
SET artifact.size=s.size
WHERE artifact.digest=s.digest;

/* artifact_reference records the child artifact referenced by parent artifact */
CREATE TABLE artifact_reference
(
  id          SERIAL PRIMARY KEY NOT NULL,
  parent_id   bigint unsigned NOT NULL,
  child_id    bigint unsigned NOT NULL,
  child_digest varchar(255) NOT NULL ,
  platform    varchar(255),
  urls        varchar(1024),
  annotations json,
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
 label_id bigint unsigned NOT NULL,
 artifact_id bigint unsigned NOT NULL,
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP,
 FOREIGN KEY (label_id) REFERENCES harbor_label(id),
 FOREIGN KEY (artifact_id) REFERENCES artifact(id),
 CONSTRAINT unique_label_reference UNIQUE (label_id,artifact_id)
);

/*move the labels added to tag to artifact*/
INSERT IGNORE INTO label_reference (label_id, artifact_id, creation_time, update_time)
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
);

/*remove the records for images in table 'harbor_resource_label'*/
DELETE FROM harbor_resource_label WHERE resource_type = 'i';

CREATE TABLE audit_log
(
 id             SERIAL PRIMARY KEY NOT NULL,
 project_id     int NOT NULL,
 operation      varchar(20) NOT NULL,
 resource_type  varchar(255) NOT NULL,
 resource       varchar(1024) NOT NULL,
 username       varchar(255) NOT NULL,
 op_time        timestamp default CURRENT_TIMESTAMP
);

/*migrate access log to audit log*/
CREATE PROCEDURE PROC_UPDATE_AUDIT_LOG ( ) BEGIN
	INSERT INTO audit_log ( project_id, operation, resource_type, resource, username, op_time ) SELECT
	access.project_id,
	access.operation,
	'project',
	access.repo_name,
	access.username,
	access.op_time
	FROM
		access_log AS access
	WHERE
		( access.operation = 'create' AND access.repo_tag = 'N/A' )
		OR ( access.operation = 'delete' AND access.repo_tag = 'N/A' );
	INSERT INTO audit_log ( project_id, operation, resource_type, resource, username, op_time ) SELECT
	access.project_id,
	'delete',
	'artifact',
	CONCAT( access.repo_name, ':', access.repo_tag ),
	access.username,
	access.op_time
	FROM
		access_log AS access
	WHERE
		access.operation = 'delete'
		AND access.repo_tag != 'N/A';
	INSERT INTO audit_log ( project_id, operation, resource_type, resource, username, op_time ) SELECT
	access.project_id,
	'create',
	'artifact',
	CONCAT( access.repo_name, ':', access.repo_tag ),
	access.username,
	access.op_time
	FROM
		access_log AS access
	WHERE
		access.operation = 'push';
	INSERT INTO audit_log ( project_id, operation, resource_type, resource, username, op_time ) SELECT
	access.project_id,
	'pull',
	'artifact',
	CONCAT( access.repo_name, ':', access.repo_tag ),
	access.username,
	access.op_time
	FROM
		access_log AS access
	WHERE
	access.operation = 'pull';
END;

CALL PROC_UPDATE_AUDIT_LOG();

/*drop access table after migrate to audit log*/
DROP TABLE IF EXISTS access_log;

/*remove the constraint for project_id in table 'notification_policy'*/
ALTER TABLE notification_policy DROP INDEX unique_project_id;

/*the existing policy has no name, to make sure the unique constraint for name works, use the id as name*/
/*if the name is set via API, it will be force to be changed with new pattern*/
UPDATE notification_policy SET name=CONCAT('policy_', id);
/*add the unique constraint for name in table 'notification_policy'*/
ALTER TABLE notification_policy ADD UNIQUE (name);

ALTER TABLE replication_task MODIFY COLUMN src_resource varchar(512) DEFAULT NULL;
ALTER TABLE replication_task MODIFY COLUMN dst_resource varchar(512) DEFAULT NULL;

/*remove count from quota hard and quota_usage used json*/
UPDATE quota SET hard = json_remove(hard, '$.count');
UPDATE quota_usage SET used = json_remove(used, '$.count');

/* make Clair and Trivy as reserved name for scanners in-tree */
UPDATE scanner_registration SET name = concat_ws('-', name, uuid) WHERE name IN ('Clair', 'Trivy') AND immutable = FALSE;
UPDATE scanner_registration SET name = SUBSTRING_INDEX(name, '-', 1) WHERE immutable = TRUE;

/*update event types in table 'notification_policy'*/
UPDATE notification_policy SET event_types = '["DOWNLOAD_CHART","DELETE_CHART","UPLOAD_CHART","DELETE_ARTIFACT","PULL_ARTIFACT","PUSH_ARTIFACT","SCANNING_FAILED","SCANNING_COMPLETED"]';

/*update event type in table 'notification_job'*/
UPDATE notification_job
SET event_type = CASE
	WHEN notification_job.event_type = 'downloadChart' THEN 'DOWNLOAD_CHART'
	WHEN notification_job.event_type = 'deleteChart' THEN 'DELETE_CHART'
	WHEN notification_job.event_type = 'uploadChart' THEN 'UPLOAD_CHART'
	WHEN notification_job.event_type = 'deleteImage' THEN 'DELETE_ARTIFACT'
	WHEN notification_job.event_type = 'pullImage' THEN 'PULL_ARTIFACT'
	WHEN notification_job.event_type = 'pushImage' THEN 'PUSH_ARTIFACT'
	WHEN notification_job.event_type = 'scanningFailed' THEN 'SCANNING_FAILED'
	WHEN notification_job.event_type = 'scanningCompleted' THEN 'SCANNING_COMPLETED'
	ELSE event_type
END;
