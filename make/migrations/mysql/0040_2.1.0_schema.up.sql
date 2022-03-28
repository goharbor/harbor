CREATE PROCEDURE PROC_ADD_COLUMN_IF_NOT_EXISTS (in TB_NAME varchar(64), in CL_NAME varchar(64), in CL_TYPE varchar(64)) BEGIN
SELECT count(*) INTO @EXIST_CL
FROM INFORMATION_SCHEMA.COLUMNS
WHERE TABLE_SCHEMA = database()
  AND TABLE_NAME = TB_NAME
  AND COLUMN_NAME = CL_NAME LIMIT 1;

SET @sql_cl = IF (@EXIST_CL <= 0, CONCAT('ALTER TABLE `', TB_NAME, '` ADD COLUMN `',CL_NAME, '` ', CL_TYPE),
    'select \' COLUMN EXISTS\' status');
PREPARE stmt_cl FROM @sql_cl;
EXECUTE stmt_cl;
DEALLOCATE PREPARE stmt_cl;
END;

CALL PROC_ADD_COLUMN_IF_NOT_EXISTS('project', 'registry_id', 'int');

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
    start_time timestamp(6) DEFAULT CURRENT_TIMESTAMP(6),
    end_time timestamp(6) NULL DEFAULT NULL,
    revision int,
    PRIMARY KEY (id),
    CHECK (extra_attrs is null or JSON_VALID (extra_attrs))
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
    creation_time timestamp(6) DEFAULT CURRENT_TIMESTAMP(6),
    start_time timestamp(6),
    update_time timestamp(6),
    end_time timestamp(6) NULL DEFAULT NULL,
    FOREIGN KEY (execution_id) REFERENCES execution(id),
    CHECK (extra_attrs is null or JSON_VALID (extra_attrs))
);

CALL PROC_ADD_COLUMN_IF_NOT_EXISTS('blob', 'update_time', 'timestamp(6) default CURRENT_TIMESTAMP(6)');
CALL PROC_ADD_COLUMN_IF_NOT_EXISTS('blob', 'status', 'varchar(255) default \'none\'');
CALL PROC_ADD_COLUMN_IF_NOT_EXISTS('blob', 'version', 'BIGINT default 0');

CREATE PROCEDURE PROC_CREATE_INDEX_IF_NOT_EXISTS (in TB_NAME varchar(64), in CL_NAME varchar(64), in IND_NAME varchar(64)) BEGIN
SELECT count(*) INTO @EXIST_IND
FROM INFORMATION_SCHEMA.STATISTICS
WHERE TABLE_SCHEMA = database()
  AND TABLE_NAME = TB_NAME
  AND INDEX_NAME = IND_NAME LIMIT 1;

SET @sql_ind = IF (@EXIST_IND <= 0, CONCAT('CREATE INDEX `', IND_NAME, '` ON `', TB_NAME, '` (', CL_NAME, ')'),
    'select \' INDEX EXISTS\' status');
PREPARE stmt_ind FROM @sql_ind;
EXECUTE stmt_ind;
DEALLOCATE PREPARE stmt_ind;
END;

CALL PROC_CREATE_INDEX_IF_NOT_EXISTS('blob', 'status', 'idx_status');
CALL PROC_CREATE_INDEX_IF_NOT_EXISTS('blob', 'version', 'idx_version');

CREATE TABLE IF NOT EXISTS p2p_preheat_instance (
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
    creation_time timestamp(6),
    update_time timestamp(6),
    UNIQUE (name, project_id)
);

CALL PROC_ADD_COLUMN_IF_NOT_EXISTS('schedule', 'vendor_type', 'varchar(16)');
CALL PROC_ADD_COLUMN_IF_NOT_EXISTS('schedule', 'vendor_id', 'int');
CALL PROC_ADD_COLUMN_IF_NOT_EXISTS('schedule', 'cron', 'varchar(64)');
CALL PROC_ADD_COLUMN_IF_NOT_EXISTS('schedule', 'callback_func_name', 'varchar(128)');
CALL PROC_ADD_COLUMN_IF_NOT_EXISTS('schedule', 'callback_func_param', 'text');

/*abstract the cron, callback function parameters from table retention_policy*/
UPDATE schedule, (
    SELECT id, replace(json_extract(data,'$.trigger.references.job_id'),'"','') AS schedule_id,
        replace(json_extract(data,'$.trigger.settings.cron'),'"','') AS cron
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

CALL PROC_ADD_COLUMN_IF_NOT_EXISTS('artifact', 'icon', 'varchar(255)');

/*remove the constraint for name in table 'notification_policy'*/
/*ALTER TABLE notification_policy DROP CONSTRAINT notification_policy_name_key;*/
/*add union unique constraint for name and project_id in table 'notification_policy'*/
ALTER TABLE notification_policy ADD UNIQUE(name,project_id);

CREATE TABLE IF NOT EXISTS data_migrations (
    id SERIAL PRIMARY KEY NOT NULL,
    version int,
    creation_time timestamp(6) default CURRENT_TIMESTAMP(6),
    update_time timestamp(6) default CURRENT_TIMESTAMP(6)
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

