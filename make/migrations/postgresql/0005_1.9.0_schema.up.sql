/*add notification policy table*/
create table notification_policy (
 id SERIAL NOT NULL,
 name varchar(256),
 project_id int NOT NULL,
 enabled boolean NOT NULL DEFAULT true,
 description text,
 targets text,
 event_types text,
 creator varchar(256),
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP,
 PRIMARY KEY (id)
 );

/*add notification job table*/
 CREATE TABLE notification_job (
 id SERIAL NOT NULL,
 policy_id int NOT NULL,
 status varchar(32),
 /* event_type is the type of trigger event, eg. pushImage, pullImage, uploadChart... */
 event_type varchar(256),
 /* notify_type is the type to notify event to user, eg. HTTP, Email... */
 notify_type varchar(256),
 job_detail text,
 job_uuid varchar(64),
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP,
 PRIMARY KEY (id)
 );
 