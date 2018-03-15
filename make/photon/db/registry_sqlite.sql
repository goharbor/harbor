create table access (
 access_id INTEGER PRIMARY KEY,
 access_code char(1),
 comment varchar (30)
);

insert into access (access_code, comment) values 
('M', 'Management access for project'),
('R', 'Read access for project'),
('W', 'Write access for project'),
('D', 'Delete access for project'),
('S', 'Search access for project');


create table role (
 role_id INTEGER PRIMARY KEY,
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


create table user (
 user_id INTEGER PRIMARY KEY,
/*
 The max length of username controlled by API is 20, 
 and 11 is reserved for marking the deleted users.
 The mark of deleted user is "#user_id".
 The 11 consist of 10 for the max value of user_id(4294967295)  
 in MySQL and 1 of '#'.
*/
 username varchar(255),
/*
 11 bytes is reserved for marking the deleted users.
*/
 email varchar(255),
 password varchar(40) NOT NULL,
 realname varchar (255) NOT NULL,
 comment varchar (30),
 deleted tinyint (1) DEFAULT 0 NOT NULL,
 reset_uuid varchar(40) DEFAULT NULL,
 salt varchar(40) DEFAULT NULL,
 sysadmin_flag tinyint (1),
 creation_time timestamp,
 update_time timestamp,
 UNIQUE (username),
 UNIQUE (email)
);

insert into user (username, email, password, realname, comment, deleted, sysadmin_flag, creation_time, update_time) values 
('admin', 'admin@example.com', '', 'system admin', 'admin user',0, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('anonymous', 'anonymous@example.com', '', 'anonymous user', 'anonymous user', 1, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

create table user_group (
id INTEGER PRIMARY KEY,
group_name varchar(255) NOT NULL,
group_type int default 0,
ldap_group_dn varchar(512) NOT NULL,
creation_time timestamp default CURRENT_TIMESTAMP,
update_time timestamp default CURRENT_TIMESTAMP
);

create table project (
 project_id INTEGER PRIMARY KEY,
 owner_id int NOT NULL,
/*
 The max length of name controlled by API is 30, 
 and 11 is reserved for marking the deleted project.
*/
 name varchar (255) NOT NULL,
 creation_time timestamp,
 update_time timestamp,
 deleted tinyint (1) DEFAULT 0 NOT NULL,
 FOREIGN KEY (owner_id) REFERENCES user(user_id),
 UNIQUE (name)
);

insert into project (owner_id, name, creation_time, update_time) values 
(1, 'library', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);



create table project_member (
 id INTEGER PRIMARY KEY,
 project_id int NOT NULL,
 entity_id int NOT NULL,
 entity_type char NOT NULL,
 role int NOT NULL,
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP,
 UNIQUE (project_id, entity_id, entity_type)
 );

insert into project_member (project_id, entity_id, role, entity_type) values
(1, 1, 1, 'u');

create table project_metadata (
 id INTEGER PRIMARY KEY,
 project_id int NOT NULL,
 name varchar(255) NOT NULL,
 value varchar(255),
 creation_time timestamp,
 update_time timestamp,
 deleted tinyint (1) DEFAULT 0 NOT NULL,
 UNIQUE(project_id, name),
 FOREIGN KEY (project_id) REFERENCES project(project_id)
);

insert into project_metadata (id, project_id, name, value, creation_time, update_time, deleted) values
(1, 1, 'public', 'true', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 0);

create table access_log (
 log_id INTEGER PRIMARY KEY,
 username varchar (255) NOT NULL,
 project_id int NOT NULL,
 repo_name varchar (256), 
 repo_tag varchar (128),
 GUID varchar(64), 
 operation varchar(20) NOT NULL,
 op_time timestamp
);

CREATE INDEX pid_optime ON access_log (project_id, op_time);

create table repository (
 repository_id INTEGER PRIMARY KEY,
 name varchar(255) NOT NULL,
 project_id int NOT NULL,
 description text,
 pull_count int DEFAULT 0 NOT NULL,
 star_count int DEFAULT 0 NOT NULL,
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP,
 UNIQUE (name)
);

create table replication_policy (
 id INTEGER PRIMARY KEY,
 name varchar(256),
 project_id int NOT NULL,
 target_id int NOT NULL,
 enabled tinyint(1) NOT NULL DEFAULT 1,
 description text,
 deleted tinyint (1) DEFAULT 0 NOT NULL,
 cron_str varchar(256),
 filters varchar(1024),
 replicate_deletion tinyint (1) DEFAULT 0 NOT NULL,
 start_time timestamp NULL,
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP
 );

create table replication_target (
 id INTEGER PRIMARY KEY,
 name varchar(64),
 url varchar(64),
 username varchar(255),
 password varchar(128),
 /*
 target_type indicates the type of target registry,
 0 means it's a harbor instance,
 1 means it's a regulart registry
 */
 target_type tinyint(1) NOT NULL DEFAULT 0,
 insecure tinyint(1) NOT NULL DEFAULT 0,
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP
 );

create table replication_job (
 id INTEGER PRIMARY KEY, 
 status varchar(64) NOT NULL,
 policy_id int NOT NULL,
 repository varchar(256) NOT NULL,
 operation  varchar(64) NOT NULL,
 tags   varchar(16384),
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP
 );

create table replication_immediate_trigger (
 id INTEGER PRIMARY KEY, 
 policy_id int NOT NULL,
 namespace varchar(256) NOT NULL,
 on_push tinyint(1) NOT NULL DEFAULT 0,
 on_deletion tinyint(1) NOT NULL DEFAULT 0,
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP
 );


create table img_scan_job (
 id INTEGER PRIMARY KEY,
 status varchar(64) NOT NULL,
 repository varchar(256) NOT NULL,
 tag   varchar(128) NOT NULL,
 digest varchar(64),
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP
 );

create table img_scan_overview (
 id INTEGER PRIMARY KEY, 
 image_digest varchar(128),
 scan_job_id int NOT NULL,
 /* 0 indicates none, the higher the number, the more severe the status */
 severity int NOT NULL default 0,
 /* the json string to store components severity status, currently use a json to be more flexible and avoid creating additional tables. */
 components_overview varchar(2048),
 /* primary key for querying details, in clair it should be the name of the "top layer" */
 details_key varchar(128),
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP,
 UNIQUE(image_digest)
 );

CREATE INDEX policy ON replication_job (policy_id);
CREATE INDEX poid_uptime ON replication_job (policy_id, update_time);
 
create table clair_vuln_timestamp (
id INTEGER PRIMARY KEY, 
namespace varchar(128) NOT NULL,
last_update timestamp NOT NULL,
UNIQUE(namespace)
);

create table properties (
 id INTEGER PRIMARY KEY,
 k varchar(64) NOT NULL,
 v varchar(128) NOT NULL,
 UNIQUE(k)
 );

create table harbor_label (
 id INTEGER PRIMARY KEY,
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
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP,
 UNIQUE(name, scope)
 );

create table harbor_resource_label (
 id INTEGER PRIMARY KEY,
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
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP,
 UNIQUE (label_id,resource_id,resource_name,resource_type)
 );

create table alembic_version (
    version_num varchar(32) NOT NULL
);

insert into alembic_version values ('1.4.0');

