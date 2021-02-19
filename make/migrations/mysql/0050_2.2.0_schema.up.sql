/*
Fixes issue https://github.com/goharbor/harbor/issues/13317
  Ensure the role_id of maintainer is 4 and the role_id of limited guest is 5
*/
UPDATE role SET role_id=4 WHERE name='maintainer' AND role_id!=4;
UPDATE role SET role_id=5 WHERE name='limitedGuest' AND role_id!=5;

/*
 Fixes issue https://github.com/goharbor/harbor/issues/12700
 Add the empty CVE allowlist to project library.
 */
INSERT INTO cve_allowlist (project_id, items) SELECT 1, '[]' WHERE NOT EXISTS (SELECT id FROM cve_allowlist WHERE project_id=1);

/*
Clean the dirty data in quota/quota_usage
  Remove quota/quota_usage when the referenced project not exists
*/
DELETE FROM quota WHERE reference='project' AND reference_id NOT IN (SELECT project_id FROM project WHERE deleted=FALSE);
DELETE FROM quota_usage WHERE reference='project' AND reference_id NOT IN (SELECT project_id FROM project WHERE deleted=FALSE);

ALTER TABLE schedule ADD COLUMN cron_type varchar(64);
ALTER TABLE robot ADD COLUMN secret varchar(2048);
ALTER TABLE robot ADD COLUMN salt varchar(64);

SET sql_mode = '';
ALTER TABLE task ADD COLUMN vendor_type varchar(16);
UPDATE task, execution SET task.vendor_type = execution.vendor_type WHERE task.execution_id = execution.id;
ALTER TABLE task MODIFY COLUMN vendor_type varchar(16) NOT NULL;

ALTER TABLE execution ADD COLUMN update_time timestamp;

UPDATE artifact AS art
SET size = ( SELECT sum( size )  FROM `blob` WHERE digest IN ( SELECT digest_blob FROM artifact_blob WHERE digest_af = art.digest ) );

ALTER TABLE robot ADD COLUMN duration int;

CREATE TABLE  IF NOT EXISTS role_permission (
 id SERIAL PRIMARY KEY NOT NULL,
 role_type varchar(255) NOT NULL,
 role_id int NOT NULL,
 permission_policy_id int NOT NULL,
 creation_time timestamp default CURRENT_TIMESTAMP,
 CONSTRAINT unique_role_permission UNIQUE (role_type, role_id, permission_policy_id)
);

CREATE TABLE  IF NOT EXISTS permission_policy (
 id SERIAL PRIMARY KEY NOT NULL,
 /*
  scope:
   system level: /system
   project level: /project/{id}
   all project: /project/ *
  */
 scope varchar(255) NOT NULL,
 resource varchar(255),
 action varchar(255),
 effect varchar(255),
 creation_time timestamp default CURRENT_TIMESTAMP,
 CONSTRAINT unique_rbac_policy UNIQUE (scope(64), resource(64), action(64), effect(64))
);

/*delete the replication execution records whose policy doesn't exist*/
DELETE re FROM replication_execution re
        LEFT JOIN replication_policy rp ON re.policy_id=rp.id
        WHERE rp.id IS NULL;

/*delete the replication task records whose execution doesn't exist*/
DELETE rt FROM replication_task rt
        LEFT JOIN replication_execution re ON rt.execution_id=re.id
        WHERE re.id IS NULL;

/*fill the task count, status and end_time of execution based on the tasks*/
CREATE PROCEDURE PROC_UPDATE_REPLICATION_EXECUTION ( ) BEGIN
	DECLARE
		rep_exec_id INT;
	DECLARE
		rep_exec_status VARCHAR ( 32 );
	DECLARE
		rep_exec_done bool DEFAULT FALSE;
	DECLARE
		rep_exec CURSOR FOR SELECT
		id,
	STATUS
	FROM
		replication_execution;
	DECLARE
		CONTINUE HANDLER FOR NOT FOUND
		SET rep_exec_done = 1;
	OPEN rep_exec;
	read_rep_exec :
	LOOP
			FETCH rep_exec INTO rep_exec_id,
			rep_exec_status;
		IF
			rep_exec_done THEN
				LEAVE read_rep_exec;

		END IF;
		IF
			rep_exec_status != 'Stopped'
			AND rep_exec_status != 'Failed'
			AND rep_exec_status != 'Succeed' THEN
			BEGIN
				DECLARE
					rep_task_status VARCHAR ( 32 );
				DECLARE
					rep_task_status_count INT;
				DECLARE
					rep_task_done bool DEFAULT FALSE;
				DECLARE
					status_count CURSOR FOR SELECT STATUS
					,
					COUNT( * ) AS c
				FROM
					replication_task
				WHERE
					execution_id = rep_exec_id
				GROUP BY
					STATUS;
				DECLARE
					CONTINUE HANDLER FOR NOT FOUND
					SET rep_task_done = 1;
				OPEN status_count;
				read_rep_task :
				LOOP
						FETCH status_count INTO rep_task_status,
						rep_task_status_count;
					IF
						rep_task_done THEN
							LEAVE read_rep_task;

					END IF;
					IF
						rep_task_status = 'Stopped' THEN
							UPDATE replication_execution
							SET stopped = rep_task_status_count
						WHERE
							id = rep_exec_id;

						ELSEIF rep_task_status = 'Failed' THEN
						UPDATE replication_execution
						SET failed = rep_task_status_count
						WHERE
							id = rep_exec_id;

						ELSEIF rep_task_status = 'Succeed' THEN
						UPDATE replication_execution
						SET succeed = rep_task_status_count
						WHERE
							id = rep_exec_id;
						ELSE UPDATE replication_execution
						SET in_progress = rep_task_status_count
						WHERE
							id = rep_exec_id;

					END IF;

				END LOOP;
				CLOSE status_count;

			END;
			UPDATE
				replication_execution
				SET STATUS =
			CASE

					WHEN in_progress > 0 THEN
					'InProgress'
					WHEN failed > 0 THEN
					'Failed'
					WHEN stopped > 0 THEN
					'Stopped' ELSE 'Succeed'
				END
				WHERE
					id = rep_exec_id;
				UPDATE replication_execution
				SET end_time = ( SELECT MAX( end_time ) FROM replication_task WHERE execution_id = 1 )
				WHERE
					id = 1
					AND ( STATUS = 'Failed' OR STATUS = 'Stopped' OR STATUS = 'Succeed' );

			END IF;

		END LOOP;
	CLOSE rep_exec;
END;

CALL PROC_UPDATE_REPLICATION_EXECUTION();

/*move the replication execution records into the new execution table*/
ALTER TABLE replication_execution ADD COLUMN new_execution_id int;
INSERT INTO execution ( vendor_type, vendor_id, status, status_message, revision, `trigger`, start_time, end_time ) SELECT
'REPLICATION',
rep_exec.policy_id,
CASE

	WHEN rep_exec.STATUS = 'InProgress' THEN
	'Running'
	WHEN rep_exec.STATUS = 'Stopped' THEN
	'Stopped'
	WHEN rep_exec.STATUS = 'Failed' THEN
	'Error'
	WHEN rep_exec.STATUS = 'Succeed' THEN
	'Success'
	END,
	rep_exec.status_text,
	0,
CASE

		WHEN rep_exec.TRIGGER = 'scheduled' THEN
		'SCHEDULE'
		WHEN rep_exec.TRIGGER = 'event_based' THEN
		'EVENT' ELSE 'MANUAL'
	END,
	rep_exec.start_time,
	rep_exec.end_time
FROM
replication_execution AS rep_exec;
UPDATE replication_execution
SET new_execution_id = ( SELECT max( id ) FROM execution WHERE vendor_id = policy_id AND vendor_type = 'REPLICATION' );

/*move the replication task records into the new task table*/
INSERT INTO task ( vendor_type, execution_id, job_id, STATUS, status_code, status_revision, run_count, extra_attrs, creation_time, start_time, update_time, end_time ) SELECT
'REPLICATION',
( SELECT new_execution_id FROM replication_execution WHERE id = rep_task.execution_id ),
rep_task.job_id,
CASE

		WHEN rep_task.STATUS = 'InProgress' THEN
		'Running'
		WHEN rep_task.STATUS = 'Stopped' THEN
		'Stopped'
		WHEN rep_task.STATUS = 'Failed' THEN
		'Error'
		WHEN rep_task.STATUS = 'Succeed' THEN
		'Success' ELSE 'Pending'
	END,
CASE

		WHEN rep_task.STATUS = 'InProgress' THEN
		2
		WHEN rep_task.STATUS = 'Stopped' THEN
		3
		WHEN rep_task.STATUS = 'Failed' THEN
		3
		WHEN rep_task.STATUS = 'Succeed' THEN
		3 ELSE 0
	END,
	rep_task.status_revision,
	1,
	CONCAT( '{"resource_type":"', rep_task.resource_type, '","source_resource":"', rep_task.src_resource, '","destination_resource":"', rep_task.dst_resource, '","operation":"', rep_task.operation, '"}' ),
	rep_task.start_time,
	rep_task.start_time,
	rep_task.end_time,
	rep_task.end_time
FROM
	replication_task AS rep_task;

DROP TABLE IF EXISTS replication_task;
DROP TABLE IF EXISTS replication_execution;

/*move the replication schedule job records into the new schedule table*/
INSERT INTO `schedule` ( vendor_type, vendor_id, cron, callback_func_name, callback_func_param, creation_time, update_time ) SELECT
'REPLICATION',
schd.policy_id,
( SELECT `trigger` ->> '$.trigger_settings.cron' FROM replication_policy WHERE id = schd.policy_id ),
'REPLICATION_CALLBACK',
schd.policy_id,
schd.creation_time,
schd.update_time
FROM
	replication_schedule_job AS schd;

INSERT INTO execution ( vendor_type, vendor_id, STATUS, revision, `trigger`, start_time, end_time ) SELECT
'SCHEDULER',
( SELECT max( id ) FROM `schedule` WHERE vendor_id = schd.policy_id AND vendor_type = 'REPLICATION' AND callback_func_name = 'REPLICATION_CALLBACK' ),
CASE

		WHEN schd.STATUS = 'stopped' THEN
		'Stopped'
		WHEN schd.STATUS = 'error' THEN
		'Error'
		WHEN schd.STATUS = 'finished' THEN
		'Success'
		WHEN schd.STATUS = 'running' THEN
		'Running'
		WHEN schd.STATUS = 'pending' THEN
		'Running'
		WHEN schd.STATUS = 'scheduled' THEN
		'Running' ELSE 'Running'
	END,
	0,
	'MANUAL',
	schd.creation_time,
	schd.update_time
FROM
	replication_schedule_job AS schd;
INSERT INTO task ( vendor_type, execution_id, job_id, STATUS, status_code, status_revision, run_count, creation_time, start_time, update_time, end_time ) SELECT
'SCHEDULER',
(
	SELECT
		id
	FROM
		execution
	WHERE
		vendor_id = ( SELECT max( id ) FROM `schedule` WHERE vendor_id = schd.policy_id AND vendor_type = 'REPLICATION' AND callback_func_name = 'REPLICATION_CALLBACK' )
	),
	schd.job_id,
CASE

		WHEN schd.STATUS = 'stopped' THEN
		'Stopped'
		WHEN schd.STATUS = 'error' THEN
		'Error'
		WHEN schd.STATUS = 'finished' THEN
		'Success'
		WHEN schd.STATUS = 'running' THEN
		'Running'
		WHEN schd.STATUS = 'pending' THEN
		'Pending'
		WHEN schd.STATUS = 'scheduled' THEN
		'Scheduled' ELSE 'Pending'
	END,
CASE

		WHEN schd.STATUS = 'stopped' THEN
		3
		WHEN schd.STATUS = 'error' THEN
		3
		WHEN schd.STATUS = 'finished' THEN
		3
		WHEN schd.STATUS = 'running' THEN
		2
		WHEN schd.STATUS = 'pending' THEN
		0
		WHEN schd.STATUS = 'scheduled' THEN
		1 ELSE 0
	END,
	0,
	1,
	schd.creation_time,
	schd.creation_time,
	schd.update_time,
	schd.update_time
FROM
	replication_schedule_job AS schd;

DROP TABLE IF EXISTS replication_schedule_job;

/* remove the clair scanner */
DELETE FROM scan_report WHERE registration_uuid = (SELECT uuid FROM scanner_registration WHERE name = 'Clair' AND immutable = TRUE);
DELETE FROM scanner_registration WHERE name = 'Clair' AND immutable = TRUE;
UPDATE scanner_registration SET is_default = TRUE WHERE name = 'Trivy' AND immutable = TRUE;


SET sql_mode = '';
ALTER TABLE execution MODIFY COLUMN vendor_type varchar(64) NOT NULL;
ALTER TABLE `schedule` MODIFY COLUMN vendor_type varchar(64) DEFAULT NULL;
ALTER TABLE `schedule` ADD COLUMN extra_attrs JSON;
ALTER TABLE task MODIFY COLUMN vendor_type varchar(64) NOT NULL;

/* Remove these columns in scan_report because execution-task pattern will handle them */
ALTER TABLE scan_report DROP COLUMN job_id;
ALTER TABLE scan_report DROP COLUMN track_id;
ALTER TABLE scan_report DROP COLUMN requester;
ALTER TABLE scan_report DROP COLUMN status;
ALTER TABLE scan_report DROP COLUMN status_code;
ALTER TABLE scan_report DROP COLUMN status_rev;
ALTER TABLE scan_report DROP COLUMN start_time;
ALTER TABLE scan_report DROP COLUMN end_time;

/*add unique for vendor_type+vendor_id to avoid dup records when updating policies*/
ALTER TABLE schedule ADD CONSTRAINT unique_schedule UNIQUE (vendor_type, vendor_id);

/*move the gc schedule job records into the new schedule table*/
INSERT INTO `schedule` ( vendor_type, vendor_id, cron, callback_func_name, callback_func_param, cron_type, extra_attrs, creation_time, update_time ) SELECT
'GARBAGE_COLLECTION',
- 1,
schd.cron_str ->> '$.cron',
'GARBAGE_COLLECTION',
( SELECT JSON_OBJECT ( 'trigger', NULL, 'deleteuntagged', schd.job_parameters -> '$.delete_untagged', 'dryrun', FALSE, 'extra_attrs', schd.job_parameters ) ),
schd.cron_str ->> '$.type',
( SELECT JSON_OBJECT ( 'delete_untagged', schd.job_parameters -> '$.delete_untagged' ) ),
schd.creation_time,
schd.update_time
FROM
	admin_job AS schd
WHERE
	job_name = 'IMAGE_GC'
	AND job_kind = 'Periodic'
	AND deleted = FALSE;
INSERT INTO execution ( vendor_type, vendor_id, STATUS, revision, `trigger`, start_time, end_time ) SELECT
'SCHEDULER',
(
	SELECT
		id
	FROM
		`schedule`
	WHERE
		vendor_type = 'GARBAGE_COLLECTION'
		AND vendor_id =- 1
		AND cron = schd.cron_str ->> '$.cron'
		AND callback_func_name = 'GARBAGE_COLLECTION'
		AND cron_type = schd.cron_str ->> '$.type'
		AND creation_time = schd.creation_time
		AND update_time = schd.update_time
	),
CASE

		WHEN schd.STATUS = 'stopped' THEN
		'Stopped'
		WHEN schd.STATUS = 'error' THEN
		'Error'
		WHEN schd.STATUS = 'finished' THEN
		'Success'
		WHEN schd.STATUS = 'running' THEN
		'Running'
		WHEN schd.STATUS = 'pending' THEN
		'Running'
		WHEN schd.STATUS = 'scheduled' THEN
		'Running' ELSE 'Running'
	END,
	0,
	'MANUAL',
	schd.creation_time,
	schd.update_time
FROM
	admin_job AS schd
WHERE
	job_name = 'IMAGE_GC'
	AND job_kind = 'Periodic'
	AND deleted = FALSE;
INSERT INTO task ( vendor_type, execution_id, job_id, STATUS, status_code, status_revision, run_count, creation_time, start_time, update_time, end_time ) SELECT
'SCHEDULER',
(
	SELECT
		id
	FROM
		execution
	WHERE
		vendor_type = 'SCHEDULER'
		AND vendor_id = (
		SELECT
			id
		FROM
			`schedule`
		WHERE
			vendor_type = 'GARBAGE_COLLECTION'
			AND vendor_id =- 1
			AND cron = schd.cron_str ->> '$.cron'
			AND callback_func_name = 'GARBAGE_COLLECTION'
			AND cron_type = schd.cron_str ->> '$.type'
			AND creation_time = schd.creation_time
			AND update_time = schd.creation_time
		)
		AND revision = 0
		AND `trigger` = 'MANUAL'
	),
	schd.job_uuid,
CASE

		WHEN schd.STATUS = 'stopped' THEN
		'Stopped'
		WHEN schd.STATUS = 'error' THEN
		'Error'
		WHEN schd.STATUS = 'finished' THEN
		'Success'
		WHEN schd.STATUS = 'running' THEN
		'Running'
		WHEN schd.STATUS = 'pending' THEN
		'Pending'
		WHEN schd.STATUS = 'scheduled' THEN
		'Scheduled' ELSE 'Pending'
	END,
CASE

		WHEN schd.STATUS = 'stopped' THEN
		3
		WHEN schd.STATUS = 'error' THEN
		3
		WHEN schd.STATUS = 'finished' THEN
		3
		WHEN schd.STATUS = 'running' THEN
		2
		WHEN schd.STATUS = 'pending' THEN
		0
		WHEN schd.STATUS = 'scheduled' THEN
		1 ELSE 0
	END,
	0,
	1,
	schd.creation_time,
	schd.creation_time,
	schd.update_time,
	schd.update_time
FROM
	admin_job AS schd
WHERE
	job_name = 'IMAGE_GC'
	AND job_kind = 'Periodic'
	AND deleted = FALSE;

/*move the gc history into the new task&execution table*/
INSERT INTO execution ( vendor_type, vendor_id, STATUS, revision, extra_attrs, `trigger`, start_time, end_time ) SELECT
'GARBAGE_COLLECTION',
- 1,
CASE

		WHEN aj.STATUS = 'stopped' THEN
		'Stopped'
		WHEN aj.STATUS = 'error' THEN
		'Error'
		WHEN aj.STATUS = 'finished' THEN
		'Success'
		WHEN aj.STATUS = 'running' THEN
		'Running'
		WHEN aj.STATUS = 'pending' THEN
		'Running'
		WHEN aj.STATUS = 'scheduled' THEN
		'Running' ELSE 'Running'
	END,
	0,
	aj.job_parameters,
	'MANUAL',
	aj.creation_time,
	aj.update_time
FROM
	admin_job AS aj
WHERE
	job_name = 'IMAGE_GC'
	AND job_kind = 'Generic'
	AND deleted = FALSE;
INSERT INTO task ( vendor_type, execution_id, job_id, STATUS, status_code, status_revision, run_count, extra_attrs, creation_time, start_time, update_time, end_time ) SELECT
'GARBAGE_COLLECTION',
(
	SELECT
		id
	FROM
		execution
	WHERE
		vendor_type = 'GARBAGE_COLLECTION'
		AND vendor_id = - 1
		AND revision = 0
		AND extra_attrs = aj.job_parameters
		AND `trigger` = 'MANUAL'
		AND start_time = aj.creation_time
		AND end_time = aj.update_time
	),
	aj.job_uuid,
CASE

		WHEN aj.STATUS = 'stopped' THEN
		'Stopped'
		WHEN aj.STATUS = 'error' THEN
		'Error'
		WHEN aj.STATUS = 'finished' THEN
		'Success'
		WHEN aj.STATUS = 'running' THEN
		'Running'
		WHEN aj.STATUS = 'pending' THEN
		'Pending'
		WHEN aj.STATUS = 'scheduled' THEN
		'Scheduled' ELSE 'Pending'
	END,
CASE

		WHEN aj.STATUS = 'stopped' THEN
		3
		WHEN aj.STATUS = 'error' THEN
		3
		WHEN aj.STATUS = 'finished' THEN
		3
		WHEN aj.STATUS = 'running' THEN
		2
		WHEN aj.STATUS = 'pending' THEN
		0
		WHEN aj.STATUS = 'scheduled' THEN
		1 ELSE 0
	END,
	0,
	1,
	cast( aj.job_parameters AS json ),
	aj.creation_time,
	aj.creation_time,
	aj.update_time,
	aj.update_time
FROM
	admin_job AS aj
WHERE
	job_name = 'IMAGE_GC'
	AND job_kind = 'Generic'
	AND deleted = FALSE;

/*move the scan all schedule records into the new schedule table*/
INSERT INTO `schedule` ( vendor_type, vendor_id, cron, callback_func_name, cron_type, creation_time, update_time ) SELECT
'SCAN_ALL',
0,
schd.cron_str ->> 'cron',
'scanAll',
schd.cron_str ->> 'type',
schd.creation_time,
schd.update_time
FROM
	admin_job AS schd
WHERE
	job_name = 'IMAGE_SCAN_ALL'
	AND job_kind = 'Periodic'
	AND deleted = FALSE;
INSERT INTO execution ( vendor_type, vendor_id, STATUS, revision, `trigger`, start_time, end_time ) SELECT
'SCHEDULER',
(
	SELECT
		id
	FROM
		`schedule`
	WHERE
		vendor_type = 'SCAN_ALL'
		AND vendor_id =0
		AND cron = schd.cron_str ->> '$.cron'
		AND callback_func_name = 'scanAll'
		AND cron_type = schd.cron_str ->> '$.type'
		AND creation_time = schd.creation_time
		AND update_time = schd.update_time
	),
CASE

		WHEN schd.STATUS = 'stopped' THEN
		'Stopped'
		WHEN schd.STATUS = 'error' THEN
		'Error'
		WHEN schd.STATUS = 'finished' THEN
		'Success'
		WHEN schd.STATUS = 'running' THEN
		'Running'
		WHEN schd.STATUS = 'pending' THEN
		'Running'
		WHEN schd.STATUS = 'scheduled' THEN
		'Running' ELSE 'Running'
	END,
	0,
	'MANUAL',
	schd.creation_time,
	schd.update_time
FROM
	admin_job AS schd
WHERE
	job_name = 'IMAGE_SCAN_ALL'
	AND job_kind = 'Periodic'
	AND deleted = FALSE;
INSERT INTO task ( vendor_type, execution_id, job_id, STATUS, status_code, status_revision, run_count, creation_time, start_time, update_time, end_time ) SELECT
'SCHEDULER',
(
	SELECT
		id
	FROM
		execution
	WHERE
		vendor_type = 'SCHEDULER'
		AND vendor_id = (
		SELECT
			id
		FROM
			`schedule`
		WHERE
		vendor_type = 'SCAN_ALL'
		AND vendor_id =0
		AND cron = schd.cron_str ->> '$.cron'
		AND callback_func_name = 'scanAll'
		AND cron_type = schd.cron_str ->> '$.type'
		AND creation_time = schd.creation_time
		AND update_time = schd.update_time
		)
		AND revision = 0
		AND `trigger` = 'MANUAL'
	),
	schd.job_uuid,
CASE

		WHEN schd.STATUS = 'stopped' THEN
		'Stopped'
		WHEN schd.STATUS = 'error' THEN
		'Error'
		WHEN schd.STATUS = 'finished' THEN
		'Success'
		WHEN schd.STATUS = 'running' THEN
		'Running'
		WHEN schd.STATUS = 'pending' THEN
		'Pending'
		WHEN schd.STATUS = 'scheduled' THEN
		'Scheduled' ELSE 'Pending'
	END,
CASE

		WHEN schd.STATUS = 'stopped' THEN
		3
		WHEN schd.STATUS = 'error' THEN
		3
		WHEN schd.STATUS = 'finished' THEN
		3
		WHEN schd.STATUS = 'running' THEN
		2
		WHEN schd.STATUS = 'pending' THEN
		0
		WHEN schd.STATUS = 'scheduled' THEN
		1 ELSE 0
	END,
	0,
	1,
	schd.creation_time,
	schd.creation_time,
	schd.update_time,
	schd.update_time
FROM
	admin_job AS schd
WHERE
	job_name = 'IMAGE_SCAN_ALL'
	AND job_kind = 'Periodic'
	AND deleted = FALSE;

/* admin_job no more needed, drop it */
DROP TABLE IF EXISTS admin_job;

/*migrate robot_token_duration from minutes to days if exist*/
UPDATE properties SET v = CONCAT(cast(v AS UNSIGNED) / 60 / 24, '') WHERE k = 'robot_token_duration';

/*
Common vulnerability reporting schema.
Github proposal link : https://github.com/goharbor/community/pull/145
*/

/*
The old scan_report not work well with report_vulnerability_record and vulnerability_record so delete them
*/
DELETE FROM scan_report;

-- --------------------------------------------------
--  Table Structure for `main.VulnerabilityRecord`
-- --------------------------------------------------
SET sql_mode='';
CREATE TABLE IF NOT EXISTS vulnerability_record (
    id serial NOT NULL PRIMARY KEY,
    cve_id text NOT NULL DEFAULT '' ,
    registration_uuid VARCHAR(64) NOT NULL DEFAULT '',
    package text NOT NULL DEFAULT '' ,
    package_version text NOT NULL DEFAULT '' ,
    package_type text NOT NULL DEFAULT '' ,
    severity text NOT NULL DEFAULT '' ,
    fixed_version text,
    urls text,
    cvss_score_v3 double precision,
    cvss_score_v2 double precision,
    cvss_vector_v3 text,
    cvss_vector_v2 text,
    description text,
    cwe_ids text,
    vendor_attributes json,
    UNIQUE (cve_id(64), registration_uuid, package(64), package_version(64)),
    CONSTRAINT fk_registration_uuid FOREIGN  KEY(registration_uuid) REFERENCES scanner_registration(uuid) ON DELETE CASCADE
);

-- --------------------------------------------------
--  Table Structure for `main.ReportVulnerabilityRecord`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS report_vulnerability_record (
    id serial NOT NULL PRIMARY KEY,
    report_uuid VARCHAR(64) NOT NULL DEFAULT '' ,
    vuln_record_id bigint unsigned NOT NULL DEFAULT 0 ,
    UNIQUE (report_uuid, vuln_record_id),
    CONSTRAINT fk_vuln_record_id FOREIGN  KEY(vuln_record_id) REFERENCES vulnerability_record(id) ON DELETE CASCADE,
    CONSTRAINT fk_report_uuid FOREIGN  KEY(report_uuid) REFERENCES scan_report(uuid) ON DELETE CASCADE
);

/*make sure the revision of execution isn't null*/
UPDATE execution SET revision=0 WHERE revision IS NULL;

/*delete the retention execution records whose policy doesn't exist*/
DELETE re FROM retention_execution AS re LEFT JOIN retention_policy rp ON re.policy_id=rp.id
WHERE rp.id IS NULL;

/*delete the replication task records whose execution doesn't exist*/
DELETE rt FROM retention_task AS rt LEFT JOIN retention_execution re ON rt.execution_id=re.id
WHERE re.id IS NULL;

/*move the replication execution records into the new execution table*/
ALTER TABLE retention_execution ADD COLUMN new_execution_id int;

CREATE PROCEDURE PROC_UPDATE_EXECUTION_AND_RETENTION_EXECUTION ( ) BEGIN
	DECLARE
		rep_exec_id BIGINT ( 20 ) UNSIGNED;
	DECLARE
		rep_exec_policy_id INT ( 11 );
	DECLARE
		rep_exec_dry_run TINYINT ( 1 );
	DECLARE
		rep_exec_trigger VARCHAR ( 20 );
	DECLARE
		rep_exec_start_time TIMESTAMP;
	DECLARE
		rep_status VARCHAR ( 32 );
	DECLARE
		rep_end_time TIMESTAMP;
	DECLARE
		new_exec_id INTEGER;
	DECLARE
		in_progress INTEGER;
	DECLARE
		failed INTEGER;
	DECLARE
		success INTEGER;
	DECLARE
		stopped INTEGER;
	DECLARE
		rep_exec_done bool DEFAULT FALSE;
	DECLARE
		rep_exec CURSOR FOR SELECT
		id,
		policy_id,
		dry_run,
		`trigger`,
		start_time
	FROM
		retention_execution
	WHERE
		new_execution_id IS NULL;
	DECLARE
		CONTINUE HANDLER FOR NOT FOUND
		SET rep_exec_done = 1;
	OPEN rep_exec;
	read_rep_exec :
	LOOP
			FETCH rep_exec INTO rep_exec_id,
			rep_exec_policy_id,
			rep_exec_dry_run,
			rep_exec_trigger,
			rep_exec_start_time;
		IF
			rep_exec_done THEN
				LEAVE read_rep_exec;

		END IF;

		SET in_progress = 0;

		SET failed = 0;

		SET success = 0;

		SET stopped = 0;
		BEGIN
			DECLARE
				rep_task_status VARCHAR ( 32 );
			DECLARE
				rep_task_status_count INT;
			DECLARE
				rep_task_done bool DEFAULT FALSE;
			DECLARE
				status_count CURSOR FOR SELECT STATUS
				,
				COUNT( * ) AS c
			FROM
				replication_task
			WHERE
				execution_id = rep_exec_id
			GROUP BY
				STATUS;
			DECLARE
				CONTINUE HANDLER FOR NOT FOUND
				SET rep_task_done = 1;
			OPEN status_count;
			read_rep_task :
			LOOP
					FETCH status_count INTO rep_task_status,
					rep_task_status_count;
				IF
					rep_task_done THEN
						LEAVE read_rep_task;

				END IF;
				IF
					rep_task_status = 'Scheduled'
					OR rep_task_status = 'Pending'
					OR rep_task_status = 'Running' THEN

						SET in_progress = in_progress + rep_task_status_count;

					ELSEIF rep_task_status = 'Stopped' THEN

					SET stopped = stopped + rep_task_status_count;

					ELSEIF rep_task_status = 'Error' THEN

					SET failed = failed + rep_task_status_count;
					ELSE
						SET success = success + rep_task_status_count;

				END IF;

			END LOOP;
			CLOSE status_count;

		END;
		IF
			in_progress > 0 THEN

				SET rep_status = 'InProgress';

			ELSEIF failed > 0 THEN

			SET rep_status = 'Failed';

			ELSEIF stopped > 0 THEN

			SET rep_status = 'Stopped';
			ELSE
				SET rep_status = 'Succeed';

		END IF;
		SELECT
			max( end_time ) INTO rep_end_time
		FROM
			retention_task
		WHERE
			execution_id = rep_exec_id;
		INSERT INTO execution ( vendor_type, vendor_id, STATUS, revision, `trigger`, start_time, end_time, extra_attrs )
		VALUES
			(
				'RETENTION',
				rep_exec_policy_id,
				rep_status,
				0,
				rep_exec_trigger,
				rep_exec_start_time,
				rep_end_time,
				CONCAT( '{"dry_run": ', CASE rep_exec.dry_run WHEN 't' THEN 'true' ELSE 'false' END, '}' )
			);
		SELECT
			new_exec_id = LAST_INSERT_ID( );
		UPDATE retention_execution
		SET new_execution_id = new_exec_id
		WHERE
			id = rep_exec_id;

	END LOOP;
CLOSE rep_exec;
END;

CALL PROC_UPDATE_EXECUTION_AND_RETENTION_EXECUTION();

/*move the replication task records into the new task table*/
INSERT INTO task (vendor_type, execution_id, job_id, status, status_code, status_revision,
                  run_count, extra_attrs, creation_time, start_time, update_time, end_time)
SELECT 'RETENTION', (SELECT new_execution_id FROM retention_execution WHERE id=rep_task.execution_id),
        rep_task.job_id, rep_task.status, rep_task.status_code, rep_task.status_revision,
        1, CONCAT('{"total":"', rep_task.total,'","retained":"', rep_task.retained,'"}'),
        rep_task.start_time, rep_task.start_time, rep_task.end_time, rep_task.end_time
FROM retention_task AS rep_task;

DROP TABLE IF EXISTS replication_task;
DROP TABLE IF EXISTS replication_execution;
