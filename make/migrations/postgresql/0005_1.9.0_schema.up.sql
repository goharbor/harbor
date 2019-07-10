/*add webhook policy table*/
create table webhook_policy (
 id SERIAL NOT NULL,
 name varchar(256),
 project_id int NOT NULL,
 enabled boolean NOT NULL DEFAULT true,
 description text,
 targets text,
 hook_types text,
 creator varchar(256),
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP,
 PRIMARY KEY (id)
 );

/*add webhook execution table*/
 CREATE TABLE webhook_execution (
 id SERIAL NOT NULL,
 policy_id int NOT NULL,
 status varchar(32),
 hook_type varchar(256),
 job_detail text,
 job_uuid varchar(64),
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP,
 PRIMARY KEY (id)
 );
 