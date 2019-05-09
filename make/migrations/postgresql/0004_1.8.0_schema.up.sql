/*add robot account table*/
CREATE TABLE robot (
 id SERIAL PRIMARY KEY NOT NULL,
 name varchar(255),
 description varchar(1024),
 project_id int,
 expiresat bigint,
 disabled boolean DEFAULT false NOT NULL,
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP,
 CONSTRAINT unique_robot UNIQUE (name, project_id)
);

CREATE TRIGGER robot_update_time_at_modtime BEFORE UPDATE ON robot FOR EACH ROW EXECUTE PROCEDURE update_update_time_at_column();

CREATE TABLE oidc_user (
 id SERIAL NOT NULL,
 user_id int NOT NULL,
 /*
 Encoded secret
  */
 secret varchar(255) NOT NULL,
  /*
 Subject and Issuer
  Subject: Subject Identifier.
  Issuer: Issuer Identifier for the Issuer of the response.
  The sub (subject) and iss (issuer) Claims, used together, are the only Claims that an RP can rely upon as a stable identifier for the End-User
 */
 subiss varchar(255) NOT NULL,
 /*
 Encoded token
  */
 token text,
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP,
 PRIMARY KEY (id),
 FOREIGN KEY (user_id) REFERENCES harbor_user(user_id),
 UNIQUE (subiss)
);

CREATE TRIGGER oidc_user_update_time_at_modtime BEFORE UPDATE ON oidc_user FOR EACH ROW EXECUTE PROCEDURE update_update_time_at_column();

/*add master role*/
INSERT INTO role (role_code, name) VALUES ('DRWS', 'master');

/*delete replication jobs whose policy has been marked as "deleted"*/
DELETE FROM replication_job AS j
USING replication_policy AS p
WHERE j.policy_id = p.id AND p.deleted = TRUE;

/*delete replication policy which has been marked as "deleted"*/
DELETE FROM replication_policy AS p
WHERE p.deleted = TRUE;

/*upgrade the replication_target to registry*/
DROP TRIGGER replication_target_update_time_at_modtime ON replication_target;
ALTER TABLE replication_target RENAME TO registry;
ALTER TABLE registry ALTER COLUMN url TYPE varchar(256);
ALTER TABLE registry ADD COLUMN credential_type varchar(16);
ALTER TABLE registry RENAME COLUMN username TO access_key;
ALTER TABLE registry RENAME COLUMN password TO access_secret;
ALTER TABLE registry ALTER COLUMN access_secret TYPE varchar(4096);
ALTER TABLE registry ADD COLUMN type varchar(32);
ALTER TABLE registry DROP COLUMN target_type;
ALTER TABLE registry ADD COLUMN description text;
ALTER TABLE registry ADD COLUMN health varchar(16);
UPDATE registry SET type='harbor';
UPDATE registry SET credential_type='basic';

/*upgrade the replication_policy*/
ALTER TABLE replication_policy ADD COLUMN creator varchar(256);
ALTER TABLE replication_policy ADD COLUMN src_registry_id int;
/*A name filter "project_name/"+double star will be merged into the filters.
if harbor is integrated with the external project service, we cannot get the project name by ID,
which means the repilcation policy will match all resources.*/
UPDATE replication_policy SET filters='[]' WHERE filters='';
UPDATE replication_policy r SET filters=( r.filters::jsonb || (SELECT CONCAT('{"type":"name","value":"', p.name,'/**"}') FROM project p WHERE p.project_id=r.project_id)::jsonb);
ALTER TABLE replication_policy RENAME COLUMN target_id TO dest_registry_id;
ALTER TABLE replication_policy ALTER COLUMN dest_registry_id DROP NOT NULL;
ALTER TABLE replication_policy ADD COLUMN dest_namespace varchar(256);
ALTER TABLE replication_policy ADD COLUMN override boolean;
UPDATE replication_policy SET override=TRUE;
ALTER TABLE replication_policy DROP COLUMN project_id;
ALTER TABLE replication_policy RENAME COLUMN cron_str TO trigger;

DROP TRIGGER replication_immediate_trigger_update_time_at_modtime ON replication_immediate_trigger;
DROP TABLE replication_immediate_trigger;

create table replication_execution (
 id SERIAL NOT NULL,
 policy_id int NOT NULL,
 status varchar(32),
 /*the status text may contain error message whose length is very long*/
 status_text text,
 total int NOT NULL DEFAULT 0,
 failed int NOT NULL DEFAULT 0,
 succeed int NOT NULL DEFAULT 0,
 in_progress int NOT NULL DEFAULT 0,
 stopped int NOT NULL DEFAULT 0,
 trigger varchar(64),
 start_time timestamp default CURRENT_TIMESTAMP,
 end_time timestamp NULL,
 PRIMARY KEY (id)
 );
CREATE INDEX execution_policy ON replication_execution (policy_id);

create table replication_task (
 id SERIAL NOT NULL,
 execution_id int NOT NULL,
 resource_type varchar(64),
 src_resource varchar(256),
 dst_resource varchar(256),
 operation varchar(32),
 job_id varchar(64),
 status varchar(32),
 start_time timestamp default CURRENT_TIMESTAMP,
 end_time timestamp NULL,
 PRIMARY KEY (id)
);
CREATE INDEX task_execution ON replication_task (execution_id);


/*migrate each replication_job record to one replication_execution and one replication_task record*/
DO $$
DECLARE
    job RECORD;
    execid integer;
BEGIN
    FOR job IN SELECT * FROM replication_job WHERE operation != 'schedule'
    LOOP
      /*insert one execution record*/
      INSERT INTO replication_execution (policy_id, start_time) VALUES (job.policy_id, job.creation_time) RETURNING id INTO execid;
      /*insert one task record
      doesn't record the tags info in "src_resource" and "dst_resource" as the length
      of the tags may longer than the capability of the column*/
      INSERT INTO replication_task (execution_id, resource_type, src_resource, dst_resource, operation, job_id, status, start_time, end_time) 
      VALUES (execid, 'image', job.repository, job.repository, job.operation, job.job_uuid, job.status, job.creation_time, job.update_time);
    END LOOP;
END $$;
UPDATE replication_task SET status='Pending' WHERE status='pending';
UPDATE replication_task SET status='InProgress' WHERE status='scheduled';
UPDATE replication_task SET status='InProgress' WHERE status='running';
UPDATE replication_task SET status='Failed' WHERE status='error';
UPDATE replication_task SET status='Succeed' WHERE status='finished';
UPDATE replication_task SET operation='copy' WHERE operation='transfer';
UPDATE replication_task SET operation='deletion' WHERE operation='delete';

/*upgrade the replication_job to replication_schedule_job*/
DELETE FROM replication_job WHERE operation != 'schedule';
ALTER TABLE replication_job RENAME COLUMN job_uuid TO job_id;
ALTER TABLE replication_job DROP COLUMN repository;
ALTER TABLE replication_job DROP COLUMN operation;
ALTER TABLE replication_job DROP COLUMN tags;
ALTER TABLE replication_job DROP COLUMN op_uuid;
DROP INDEX policy;
DROP INDEX poid_uptime;
DROP INDEX poid_status;
DROP TRIGGER replication_job_update_time_at_modtime ON replication_job;
ALTER TABLE replication_job RENAME TO replication_schedule_job;

/*
migrate scan all schedule

If user set the scan all schedule, move it into table admin_job, and let the api the parse the json data.
*/
DO $$
BEGIN
    IF exists(select * FROM properties WHERE k = 'scan_all_policy') then
        INSERT INTO admin_job (job_name, job_kind, cron_str, status) VALUES ('IMAGE_SCAN_ALL', 'Periodic', (select v FROM properties WHERE k = 'scan_all_policy'), 'pending');
        DELETE FROM properties WHERE k='scan_all_policy';
    END IF;
END $$;


