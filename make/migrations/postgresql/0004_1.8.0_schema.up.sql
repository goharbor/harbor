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
 secret varchar(255) NOT NULL,
  /*
 Subject and Issuer
  Subject: Subject Identifier.
  Issuer: Issuer Identifier for the Issuer of the response.
  The sub (subject) and iss (issuer) Claims, used together, are the only Claims that an RP can rely upon as a stable identifier for the End-User
 */
 subiss varchar(255) NOT NULL,
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP,
 PRIMARY KEY (id),
 UNIQUE (subiss)
);

CREATE TRIGGER odic_user_update_time_at_modtime BEFORE UPDATE ON oidc_user FOR EACH ROW EXECUTE PROCEDURE update_update_time_at_column();

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
UPDATE registry SET credential_type='basic' WHERE credential_type='';
ALTER TABLE registry RENAME COLUMN username TO access_key;
ALTER TABLE registry RENAME COLUMN password TO access_secret;
ALTER TABLE registry ALTER COLUMN access_secret TYPE varchar(1024);
ALTER TABLE registry ADD COLUMN type varchar(32);
UPDATE registry SET type='harbor' WHERE type='';
ALTER TABLE registry DROP COLUMN target_type;
ALTER TABLE registry ADD COLUMN description text;
ALTER TABLE registry ADD COLUMN health varchar(16);

/*upgrade the replication_policy*/
ALTER TABLE replication_policy ADD COLUMN creator varchar(256);
ALTER TABLE replication_policy ADD COLUMN src_registry_id int;
ALTER TABLE replication_policy ADD COLUMN src_namespaces varchar(256);
/*if harbor is integrated with the external project service, the src_namespaces will be empty,
which means the repilcation policy cannot work as expected*/
UPDATE replication_policy r SET src_namespaces=(SELECT p.name FROM project p WHERE p.project_id=r.project_id);
ALTER TABLE replication_policy RENAME COLUMN target_id TO dest_registry_id;
ALTER TABLE replication_policy ALTER COLUMN dest_registry_id DROP NOT NULL;
ALTER TABLE replication_policy ADD COLUMN dest_namespace varchar(256);
ALTER TABLE replication_policy ADD COLUMN override boolean;
ALTER TABLE replication_policy DROP COLUMN project_id;

DROP TRIGGER replication_immediate_trigger_update_time_at_modtime ON replication_immediate_trigger;
DROP TABLE replication_immediate_trigger;

create table replication_execution (
 id SERIAL NOT NULL,
 policy_id int NOT NULL,
 status varchar(32),
 status_text varchar(256),
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

create table replication_schedule_job (
 id SERIAL NOT NULL,
 policy_id int NOT NULL,
 job_id varchar(64),
 status varchar(32),
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp NULL,
 PRIMARY KEY (id)
);
CREATE INDEX replication_schedule_job_index ON replication_schedule_job (policy_id);

/*
 * TODO
 * consider how to handle the replication_job;
 * the replication_job contains schedule job;
 * the schedule job has been removed from jobservice, how to handle this?
 * keep consistent with the webhook handler?
 */