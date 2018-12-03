ALTER TABLE properties ALTER COLUMN v TYPE varchar(1024);
DELETE FROM properties where k='scan_all_policy';

create table job_log (
 log_id SERIAL NOT NULL,
 job_uuid varchar (64) NOT NULL,
 creation_time timestamp default CURRENT_TIMESTAMP,
 content text,
 primary key (log_id)
);

CREATE UNIQUE INDEX job_log_uuid ON job_log (job_uuid);

ALTER TABLE replication_policy ADD CONSTRAINT unique_policy_name UNIQUE (name);
ALTER TABLE replication_target ADD CONSTRAINT unique_target_name UNIQUE (name);
