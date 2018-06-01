CREATE DATABASE registry ENCODING 'UTF8';

\c registry;

create table access (
 access_id int PRIMARY KEY NOT NULL,
 access_code char(1),
 comment varchar (30)
);

create table role (
 role_id int PRIMARY KEY NOT NULL,
 role_mask int DEFAULT 0 NOT NULL,
 role_code varchar(20),
 name varchar (20)
);

/*
role mask is used for future enhancement when a project member can have multi-roles
currently set to 0
*/

create table harbor_user (
 user_id int PRIMARY KEY NOT NULL,
 username varchar(255),
 email varchar(255),
 password varchar(40) NOT NULL,
 realname varchar (255) NOT NULL,
 comment varchar (30),
 deleted smallint DEFAULT 0 NOT NULL,
 reset_uuid varchar(40) DEFAULT NULL,
 salt varchar(40) DEFAULT NULL,
 sysadmin_flag smallint DEFAULT 0 NOT NULL,
 creation_time timestamp(0),
 update_time timestamp(0),
 UNIQUE (username),
 UNIQUE (email)
);

create table project (
 project_id int PRIMARY KEY NOT NULL,
 owner_id int NOT NULL,
 /*
 The max length of name controlled by API is 30, 
 and 11 is reserved for marking the deleted project.
 */
 name varchar (255) NOT NULL,
 creation_time timestamp,
 update_time timestamp,
 deleted smallint DEFAULT 0 NOT NULL,
 /*
 FOREIGN KEY (owner_id) REFERENCES harbor_user(user_id),
 */
 UNIQUE (name)
);

create table project_member (
 id int NOT NULL,
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

create table project_metadata (
 id int NOT NULL,
 project_id int NOT NULL,
 name varchar(255) NOT NULL,
 value varchar(255),
 creation_time timestamp default 'now'::timestamp,
 update_time timestamp default 'now'::timestamp,
 deleted smallint DEFAULT 0 NOT NULL,
 PRIMARY KEY (id),
 CONSTRAINT unique_project_id_and_name UNIQUE (project_id,name)
 /*
 FOREIGN KEY (project_id) REFERENCES project(project_id)
 */
);

CREATE TRIGGER project_metadata_update_time_at_modtime BEFORE UPDATE ON project_metadata FOR EACH ROW EXECUTE PROCEDURE update_update_time_at_column();

create table user_group (
 id int NOT NULL,
 group_name varchar(255) NOT NULL,
 group_type smallint default 0,
 ldap_group_dn varchar(512) NOT NULL,
 creation_time timestamp default 'now'::timestamp,
 update_time timestamp default 'now'::timestamp,
 PRIMARY KEY (id)
);

CREATE TRIGGER user_group_update_time_at_modtime BEFORE UPDATE ON user_group FOR EACH ROW EXECUTE PROCEDURE update_update_time_at_column();

create table access_log (
 log_id int NOT NULL,
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
 repository_id int NOT NULL,
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
 id int NOT NULL,
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
 id int NOT NULL,
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
 id int NOT NULL,
 status varchar(64) NOT NULL,
 policy_id int NOT NULL,
 repository varchar(256) NOT NULL,
 operation  varchar(64) NOT NULL,
 tags   varchar(16384),
  /*
New job service only records uuid, for compatibility in this table both IDs are stored.
 */
 job_uuid varchar(64),
 creation_time timestamp default 'now'::timestamp,
 update_time timestamp default 'now'::timestamp,
 PRIMARY KEY (id)
 );
 
CREATE INDEX policy ON replication_job (policy_id);
CREATE INDEX poid_uptime ON replication_job (policy_id, update_time);
CREATE INDEX poid_status ON replication_job (policy_id, status);
 
CREATE TRIGGER replication_job_update_time_at_modtime BEFORE UPDATE ON replication_job FOR EACH ROW EXECUTE PROCEDURE update_update_time_at_column();

create table replication_immediate_trigger (
 id int NOT NULL,
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
 id int NOT NULL,
 status varchar(64) NOT NULL,
 repository varchar(256) NOT NULL,
 tag varchar(128) NOT NULL,
 digest varchar(128),
/*
New job service only records uuid, for compatibility in this table both IDs are stored.
*/
 job_uuid varchar(64),
 creation_time timestamp default 'now'::timestamp,
 update_time timestamp default 'now'::timestamp,
 PRIMARY KEY (id)
 );

CREATE INDEX idx_status ON img_scan_job (status);
CREATE INDEX idx_digest ON img_scan_job (digest);
CREATE INDEX idx_uuid ON img_scan_job (job_uuid);
CREATE INDEX idx_repository_tag ON img_scan_job (repository,tag);
 
CREATE TRIGGER img_scan_job_update_time_at_modtime BEFORE UPDATE ON img_scan_job FOR EACH ROW EXECUTE PROCEDURE update_update_time_at_column();

create table img_scan_overview (
 id int NOT NULL,
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
id int NOT NULL, 
namespace varchar(128) NOT NULL,
last_update timestamp NOT NULL,
PRIMARY KEY(id),
UNIQUE(namespace)
);

create table properties (
 id int NOT NULL,
 k varchar(64) NOT NULL,
 v varchar(128) NOT NULL,
 PRIMARY KEY(id),
 UNIQUE (k)
 );

create table harbor_label (
 id int NOT NULL,
 name varchar(128) NOT NULL,
 description text,
 color varchar(16),
/*
's' for system level labels
'u' for user level labels
*/
 level char(1) NOT NULL,
/*
'g' for global labels
'p' for project labels
*/
 scope char(1) NOT NULL,
 project_id int,
 creation_time timestamp default 'now'::timestamp,
 update_time timestamp default 'now'::timestamp,
 PRIMARY KEY(id),
 CONSTRAINT unique_name_and_scope UNIQUE (name,scope,project_id)
 );

CREATE TRIGGER harbor_label_update_time_at_modtime BEFORE UPDATE ON harbor_label FOR EACH ROW EXECUTE PROCEDURE update_update_time_at_column();

create table harbor_resource_label (
 id int NOT NULL,
 label_id int NOT NULL,
/*
 the resource_id is the ID of project when the resource_type is p
 the resource_id is the ID of repository when the resource_type is r
*/
 resource_id int,
/*
the resource_name is the name of image when the resource_type is i
*/
 resource_name varchar(256),
/*
 'p' for project
 'r' for repository
 'i' for image
*/
 resource_type char(1) NOT NULL,
 creation_time timestamp default 'now'::timestamp,
 update_time timestamp default 'now'::timestamp,
 PRIMARY KEY(id),
 CONSTRAINT unique_label_resource UNIQUE (label_id,resource_id, resource_name, resource_type)
 );

CREATE TRIGGER harbor_resource_label_update_time_at_modtime BEFORE UPDATE ON harbor_resource_label FOR EACH ROW EXECUTE PROCEDURE update_update_time_at_column();
 
CREATE TABLE IF NOT EXISTS alembic_version (
    version_num varchar(32) NOT NULL
);