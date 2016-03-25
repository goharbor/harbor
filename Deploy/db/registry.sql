drop database if exists registry;
create database registry charset = utf8;

use registry;

create table access (
 access_id int NOT NULL AUTO_INCREMENT,
 access_code char(1),
 comment varchar (30),
 primary key (access_id)
);

insert into access values 
( null, 'M', 'Management access for project'),
( null, 'R', 'Read access for project'),
( null, 'W', 'Write access for project'),
( null, 'D', 'Delete access for project'),
( null, 'S', 'Search access for project');


create table role (
 role_id int NOT NULL AUTO_INCREMENT,
 role_mask int DEFAULT 0 NOT NULL,
 role_code varchar(20),
 name varchar (20),
 primary key (role_id)
);
/*
role mask is used for future enhancement when a project member can have multi-roles
currently set to 0
*/

insert into role values 
( null, 0, 'MDRWS', 'projectAdmin'),
( null, 0, 'RWS', 'developer'),
( null, 0, 'RS', 'guest');


create table user (
 user_id int NOT NULL AUTO_INCREMENT,
 username varchar(15),
 email varchar(30),
 password varchar(40) NOT NULL,
 realname varchar (20) NOT NULL,
 comment varchar (30),
 deleted tinyint (1) DEFAULT 0 NOT NULL,
 reset_uuid varchar(40) DEFAULT NULL,
 salt varchar(40) DEFAULT NULL,
 sysadmin_flag tinyint (1),
 creation_time timestamp,
 update_time timestamp,
 primary key (user_id),
 UNIQUE (username),
 UNIQUE (email)
);

insert into user values 
(1, 'admin', 'admin@example.com', '', 'system admin', 'admin user',0, null, '', 1, NOW(), NOW()),
(2, 'anonymous', 'anonymous@example.com', '', 'anonymous user', 'anonymous user', 1, null, '', 0, NOW(), NOW());
                                                                          
create table project (
 project_id int NOT NULL AUTO_INCREMENT,
 owner_id int NOT NULL,
 name varchar (30) NOT NULL,
 creation_time timestamp,
 update_time timestamp,
 deleted tinyint (1) DEFAULT 0 NOT NULL,
 public tinyint (1) DEFAULT 0 NOT NULL,
 primary key (project_id),
 FOREIGN KEY (owner_id) REFERENCES user(user_id),
 UNIQUE (name)
);

insert into project values 
(null, 1, 'library', NOW(), NOW(), 0, 1);

create table project_member (
 project_id int NOT NULL,
 user_id int NOT NULL,
 role int NOT NULL,
 creation_time timestamp,
 update_time timestamp,
 PRIMARY KEY (project_id, user_id),
 FOREIGN KEY (role) REFERENCES role(role_id),
 FOREIGN KEY (project_id) REFERENCES project(project_id),
 FOREIGN KEY (user_id) REFERENCES user(user_id)
 );

insert into project_member values
(1, 1, 1, NOW(), NOW());

create table access_log (
 log_id int NOT NULL AUTO_INCREMENT,
 user_id int NOT NULL,
 project_id int NOT NULL,
 repo_name varchar (40), 
 GUID varchar(64), 
 operation varchar(20) NOT NULL,
 op_time timestamp,
 primary key (log_id),
 FOREIGN KEY (user_id) REFERENCES user(user_id),
 FOREIGN KEY (project_id) REFERENCES project (project_id)
);
