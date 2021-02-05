ALTER TABLE project ADD COLUMN registry_id int;
ALTER TABLE cve_whitelist RENAME TO cve_allowlist;
UPDATE role SET name='maintainer' WHERE name='master';
UPDATE project_metadata SET name='reuse_sys_cve_allowlist' WHERE name='reuse_sys_cve_whitelist';

CREATE TABLE IF NOT EXISTS execution (
    id SERIAL NOT NULL,
    vendor_type varchar(16) NOT NULL,
    vendor_id int,
    status varchar(16),
    status_message text,
    `trigger` varchar(16) NOT NULL,
    extra_attrs JSON,
    start_time timestamp DEFAULT CURRENT_TIMESTAMP,
    end_time timestamp,
    revision int,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS task (
    id SERIAL PRIMARY KEY NOT NULL,
    execution_id bigint unsigned NOT NULL,
    job_id varchar(64),
    status varchar(16) NOT NULL,
    status_code int NOT NULL,
    status_revision int,
    status_message text,
    run_count int,
    extra_attrs JSON,
    creation_time timestamp DEFAULT CURRENT_TIMESTAMP,
    start_time timestamp,
    update_time timestamp,
    end_time timestamp,
    FOREIGN KEY (execution_id) REFERENCES execution(id)
);

ALTER TABLE `blob` ADD COLUMN update_time timestamp default CURRENT_TIMESTAMP;
ALTER TABLE `blob` ADD COLUMN status varchar(255) default 'none';
ALTER TABLE `blob` ADD COLUMN version BIGINT default 0;
CREATE INDEX idx_status ON `blob` (status);
CREATE INDEX idx_version ON `blob` (version);

CREATE TABLE p2p_preheat_instance (
  id          SERIAL PRIMARY KEY NOT NULL,
  name        varchar(255) NOT NULL,
  description varchar(255),
  vendor	  varchar(255) NOT NULL,
  endpoint    varchar(255) NOT NULL,
  auth_mode   varchar(255),
  auth_data   text,
  enabled     boolean,
  is_default  boolean,
  insecure    boolean,
  setup_timestamp int,
  UNIQUE (name)
);

CREATE TABLE IF NOT EXISTS p2p_preheat_policy (
    id SERIAL PRIMARY KEY NOT NULL,
    name varchar(255) NOT NULL,
    description varchar(1024),
    project_id int NOT NULL,
    provider_id int NOT NULL,
    filters varchar(1024),
    `trigger` varchar(255),
    enabled boolean,
    creation_time timestamp,
    update_time timestamp,
    UNIQUE (name, project_id)
);

ALTER TABLE schedule ADD COLUMN vendor_type varchar(16);
ALTER TABLE schedule ADD COLUMN vendor_id int;
ALTER TABLE schedule ADD COLUMN cron varchar(64);
ALTER TABLE schedule ADD COLUMN callback_func_name varchar(128);
ALTER TABLE schedule ADD COLUMN callback_func_param text;

/*abstract the cron, callback function parameters from table retention_policy*/
UPDATE schedule, (
    SELECT id, data->>'$.trigger.references.job_id' AS schedule_id,
        data->>'$.trigger.settings.cron' AS cron
        FROM retention_policy
    ) AS retention
SET vendor_type= 'RETENTION', vendor_id=retention.id, schedule.cron = retention.cron,
    callback_func_name = 'RETENTION', callback_func_param=concat('{"PolicyID":', retention.id, ',"Trigger":"Schedule"}')
WHERE schedule.id=retention.schedule_id;

/*create new execution and task record for each schedule*/
CREATE PROCEDURE PROC_UPDATE_EXECUTION_TASK ( ) BEGIN
INSERT INTO execution ( vendor_type, vendor_id, `trigger` ) SELECT
'SCHEDULER',
sched.id,
'MANUAL'
FROM
	`schedule` AS sched;
INSERT INTO task ( execution_id, job_id, STATUS, status_code, status_revision, run_count ) SELECT
exec.id,
sched.job_id,
sched.STATUS,
CASE
	WHEN sched.STATUS = 'Pending' THEN
	0
	WHEN sched.STATUS = 'Scheduled' THEN
	1
	WHEN sched.STATUS = 'Running' THEN
	2
	WHEN sched.STATUS = 'Stopped'
	OR sched.STATUS = 'Error'
	OR sched.STATUS = 'Success' THEN
	3 ELSE 0
END AS status_code,
	0,
	0
FROM
	`schedule` AS sched
	LEFT JOIN execution AS exec ON exec.vendor_id = sched.id
	AND exec.vendor_type = 'SCHEDULER'
	AND `trigger` = 'MANUAL';
END;

CALL PROC_UPDATE_EXECUTION_TASK();

ALTER TABLE schedule DROP COLUMN job_id;
ALTER TABLE schedule DROP COLUMN status;

UPDATE registry SET type = 'quay' WHERE type = 'quay-io';


ALTER TABLE artifact ADD COLUMN icon varchar(255);

/*remove the constraint for name in table 'notification_policy'*/
/*ALTER TABLE notification_policy DROP CONSTRAINT notification_policy_name_key;*/
/*add union unique constraint for name and project_id in table 'notification_policy'*/
ALTER TABLE notification_policy ADD UNIQUE(name,project_id);

CREATE TABLE IF NOT EXISTS data_migrations (
    id SERIAL PRIMARY KEY NOT NULL,
    version int,
    creation_time timestamp default CURRENT_TIMESTAMP,
    update_time timestamp default CURRENT_TIMESTAMP
);
INSERT INTO data_migrations (version) VALUES (
    CASE
        /*if the "extra_attrs" isn't null, it means that the deployment upgrades from v2.0*/
        WHEN (SELECT Count(*) FROM artifact WHERE extra_attrs!='')>0 THEN 30
        ELSE 0
    END
);
ALTER TABLE schema_migrations DROP COLUMN data_version;

UPDATE artifact
SET icon=(
CASE
    WHEN type='IMAGE' THEN
        'sha256:0048162a053eef4d4ce3fe7518615bef084403614f8bca43b40ae2e762e11e06'
    WHEN type='CHART' THEN
        'sha256:61cf3a178ff0f75bf08a25d96b75cf7355dc197749a9f128ed3ef34b0df05518'
    WHEN type='CNAB' THEN
        'sha256:089bdda265c14d8686111402c8ad629e8177a1ceb7dcd0f7f39b6480f623b3bd'
    ELSE
        'sha256:da834479c923584f4cbcdecc0dac61f32bef1d51e8aae598cf16bd154efab49f'
END);

