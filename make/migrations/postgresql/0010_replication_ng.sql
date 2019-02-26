-- DDL file for replication next generation

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
CREATE INDEX policy ON replication_execution (policy_id);

create table replication_task (
 id SERIAL NOT NULL,
 execution_id int NOT NULL,
 resource_type varchar(64),
 src_resource varchar(256),
 dst_resource varchar(256),
 job_id varchar(64),
 status varchar(32),
 start_time timestamp default CURRENT_TIMESTAMP,
 end_time timestamp NULL,
 PRIMARY KEY (id)
);
CREATE INDEX execution ON replication_task (execution_id);
