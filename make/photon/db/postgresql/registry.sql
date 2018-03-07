CREATE DATABASE registry ENCODING 'UTF8';

\c registry;

create table access (
 access_id SERIAL PRIMARY KEY NOT NULL,
 access_code char(1),
 comment varchar (30)
);

create table role (
 role_id SERIAL PRIMARY KEY NOT NULL,
 role_mask int DEFAULT 0 NOT NULL,
 role_code varchar(20),
 name varchar (20)
);

/*
role mask is used for future enhancement when a project member can have multi-roles
currently set to 0
*/

insert into role (role_code, name) values 
('MDRWS', 'projectAdmin'),
('RWS', 'developer'),
('RS', 'guest');

create table harbor_user (
 user_id SERIAL PRIMARY KEY NOT NULL,
 username varchar(255),
 email varchar(255),
 password varchar(40) NOT NULL,
 realname varchar (255) NOT NULL,
 comment varchar (30),
 deleted smallint DEFAULT 0 NOT NULL,
 reset_uuid varchar(40) DEFAULT NULL,
 salt varchar(40) DEFAULT NULL,
 sysadmin_flag smallint,
 creation_time timestamp(0),
 update_time timestamp(0),
 UNIQUE (username),
 UNIQUE (email)
);

insert into harbor_user (username, email, password, realname, comment, deleted, sysadmin_flag, creation_time, update_time) values 
('admin', 'admin@example.com', '', 'system admin', 'admin user',0, 1, NOW(), NOW()),
('anonymous', 'anonymous@example.com', '', 'anonymous user', 'anonymous user', 1, 0, NOW(), NOW());

create table project (
 project_id SERIAL PRIMARY KEY NOT NULL,
 owner_id int NOT NULL,
 name varchar (255) NOT NULL,
 creation_time timestamp,
 update_time timestamp,
 deleted smallint DEFAULT 0 NOT NULL,
 FOREIGN KEY (owner_id) REFERENCES harbor_user(user_id),
 UNIQUE (name)
);

insert into project (owner_id, name, creation_time, update_time) values 
(1, 'library', NOW(), NOW());

create table project_member (
 id SERIAL NOT NULL,
 project_id int NOT NULL,
 entity_id int NOT NULL,
 /*
 entity_type indicates the type of member,
 u for user, g for user group
 */
 entity_type char(1) NOT NULL,
 role int NOT NULL,
 creation_time timestamp default 'now'::timestamp,
 update_time timestamp default 'now'::timestamp,
 PRIMARY KEY (id),
 CONSTRAINT unique_project_entity_type UNIQUE (project_id, entity_id, entity_type)
);

CREATE FUNCTION update_update_time_at_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
  BEGIN
    NEW.update_time = NOW();
    RETURN NEW;
  END;
$$;

CREATE TRIGGER project_member_update_time_at_modtime BEFORE UPDATE ON project_member FOR EACH ROW EXECUTE PROCEDURE update_update_time_at_column();

insert into project_member (project_id, entity_id, role, entity_type) values
(1, 1, 1, 'u');

create table project_metadata (
 id SERIAL NOT NULL,
 project_id int NOT NULL,
 name varchar(255) NOT NULL,
 value varchar(255),
 creation_time timestamp default 'now'::timestamp,
 update_time timestamp default 'now'::timestamp,
 deleted smallint DEFAULT 0 NOT NULL,
 PRIMARY KEY (id),
 CONSTRAINT unique_project_id_and_name UNIQUE (project_id,name),
 FOREIGN KEY (project_id) REFERENCES project(project_id)
);

CREATE TRIGGER project_metadata_update_time_at_modtime BEFORE UPDATE ON project_metadata FOR EACH ROW EXECUTE PROCEDURE update_update_time_at_column();

insert into project_metadata (project_id, name, value, creation_time, update_time, deleted) values
(1, 'public', 'true', NOW(), NOW(), 0);

create table user_group (
 id SERIAL NOT NULL,
 group_name varchar(255) NOT NULL,
 group_type smallint default 0,
 group_property varchar(512) NOT NULL,
 creation_time timestamp default 'now'::timestamp,
 update_time timestamp default 'now'::timestamp,
 PRIMARY KEY (id)
);

CREATE TRIGGER user_group_update_time_at_modtime BEFORE UPDATE ON user_group FOR EACH ROW EXECUTE PROCEDURE update_update_time_at_column();

create table access_log (
 log_id SERIAL NOT NULL,
 username varchar (255) NOT NULL,
 project_id int NOT NULL,
 repo_name varchar (256), 
 repo_tag varchar (128),
 GUID varchar(64), 
 operation varchar(20) NOT NULL,
 op_time timestamp,
 primary key (log_id)
);

CREATE INDEX pid_optime ON access_log (project_id, op_time);

create table repository (
 repository_id SERIAL NOT NULL,
 name varchar(255) NOT NULL,
 project_id int NOT NULL,
 description text,
 pull_count int DEFAULT 0 NOT NULL,
 star_count int DEFAULT 0 NOT NULL,
 creation_time timestamp default 'now'::timestamp,
 update_time timestamp default 'now'::timestamp,
 primary key (repository_id),
 UNIQUE (name)
);

CREATE TRIGGER repository_update_time_at_modtime BEFORE UPDATE ON repository FOR EACH ROW EXECUTE PROCEDURE update_update_time_at_column();

create table replication_policy (
 id SERIAL NOT NULL,
 name varchar(256),
 project_id int NOT NULL,
 target_id int NOT NULL,
 enabled SMALLINT NOT NULL DEFAULT 1,
 description text,
 deleted SMALLINT DEFAULT 0 NOT NULL,
 cron_str varchar(256),
 filters varchar(1024),
 replicate_deletion SMALLINT DEFAULT 0 NOT NULL,
 start_time timestamp NULL,
 creation_time timestamp default 'now'::timestamp,
 update_time timestamp default 'now'::timestamp,
 PRIMARY KEY (id)
 );

CREATE TRIGGER replication_policy_update_time_at_modtime BEFORE UPDATE ON replication_policy FOR EACH ROW EXECUTE PROCEDURE update_update_time_at_column();

create table replication_target (
 id SERIAL NOT NULL,
 name varchar(64),
 url varchar(64),
 username varchar(255),
 password varchar(128),
 /*
 target_type indicates the type of target registry,
 0 means it's a harbor instance,
 1 means it's a regulart registry
 */
 target_type SMALLINT NOT NULL DEFAULT 0,
 insecure SMALLINT NOT NULL DEFAULT 0,
 creation_time timestamp default 'now'::timestamp,
 update_time timestamp default 'now'::timestamp,
 PRIMARY KEY (id)
 );
 
CREATE TRIGGER replication_target_update_time_at_modtime BEFORE UPDATE ON replication_target FOR EACH ROW EXECUTE PROCEDURE update_update_time_at_column();

create table replication_job (
 id SERIAL NOT NULL,
 status varchar(64) NOT NULL,
 policy_id int NOT NULL,
 repository varchar(256) NOT NULL,
 operation  varchar(64) NOT NULL,
 tags   varchar(16384),
 creation_time timestamp default 'now'::timestamp,
 update_time timestamp default 'now'::timestamp,
 PRIMARY KEY (id)
 );
 
CREATE INDEX policy ON replication_job (policy_id);
CREATE INDEX poid_uptime ON replication_job (policy_id, update_time);
 
CREATE TRIGGER replication_job_update_time_at_modtime BEFORE UPDATE ON replication_job FOR EACH ROW EXECUTE PROCEDURE update_update_time_at_column();

create table replication_immediate_trigger (
 id SERIAL NOT NULL,
 policy_id int NOT NULL,
 namespace varchar(256) NOT NULL,
 on_push SMALLINT NOT NULL DEFAULT 0,
 on_deletion SMALLINT NOT NULL DEFAULT 0,
 creation_time timestamp default 'now'::timestamp,
 update_time timestamp default 'now'::timestamp,
 PRIMARY KEY (id)
 );
 
 CREATE TRIGGER replication_immediate_trigger_update_time_at_modtime BEFORE UPDATE ON replication_immediate_trigger FOR EACH ROW EXECUTE PROCEDURE update_update_time_at_column();

 create table img_scan_job (
 id SERIAL NOT NULL,
 status varchar(64) NOT NULL,
 repository varchar(256) NOT NULL,
 tag varchar(128) NOT NULL,
 digest varchar(128),
 creation_time timestamp default 'now'::timestamp,
 update_time timestamp default 'now'::timestamp,
 PRIMARY KEY (id)
 );
 
CREATE TRIGGER img_scan_job_update_time_at_modtime BEFORE UPDATE ON img_scan_job FOR EACH ROW EXECUTE PROCEDURE update_update_time_at_column();

create table img_scan_overview (
 id SERIAL NOT NULL,
 image_digest varchar(128) NOT NULL,
 scan_job_id int NOT NULL,
 /* 0 indicates none, the higher the number, the more severe the status */
 severity int NOT NULL default 0,
 /* the json string to store components severity status, currently use a json to be more flexible and avoid creating additional tables. */
 components_overview varchar(2048),
 /* primary key for querying details, in clair it should be the name of the "top layer" */
 details_key varchar(128),
 creation_time timestamp default 'now'::timestamp,
 update_time timestamp default 'now'::timestamp,
 PRIMARY KEY(id),
 UNIQUE(image_digest)
 );
 
CREATE TRIGGER img_scan_overview_update_time_at_modtime BEFORE UPDATE ON img_scan_overview FOR EACH ROW EXECUTE PROCEDURE update_update_time_at_column();

create table clair_vuln_timestamp (
id SERIAL NOT NULL, 
namespace varchar(128) NOT NULL,
last_update timestamp NOT NULL,
PRIMARY KEY(id),
UNIQUE(namespace)
);

create table properties (
 id SERIAL NOT NULL,
 k varchar(64) NOT NULL,
 v varchar(128) NOT NULL,
 PRIMARY KEY(id),
 UNIQUE (k)
 );
 
CREATE TABLE IF NOT EXISTS alembic_version (
    version_num varchar(32) NOT NULL
);

insert into alembic_version values ('1.4.0');

