/* add table for CVE whitelist */
CREATE TABLE cve_whitelist (
    id SERIAL PRIMARY KEY NOT NULL,
    project_id int,
    creation_time timestamp default CURRENT_TIMESTAMP,
    update_time timestamp default CURRENT_TIMESTAMP,
    expires_at bigint,
    items text NOT NULL,
    UNIQUE (project_id)
);

create table retention_policy
(
	id serial PRIMARY KEY NOT NULL,
	scope_level varchar(20),
	scope_reference integer,
	trigger_kind varchar(20),
	data text,
	create_time time,
	update_time time
);

create table retention_execution
(
	id integer PRIMARY KEY NOT NULL,
	policy_id integer,
	status varchar(20),
	status_text text,
	dry boolean,
	trigger varchar(20),
	total integer,
	succeed integer,
	failed integer,
	in_progress integer,
	stopped integer,
	start_time time,
	end_time time
);

create table retention_task
(
	id SERIAL NOT NULL,
	execution_id integer,
	status varchar(32),
	start_time timestamp default CURRENT_TIMESTAMP,
	end_time timestamp default CURRENT_TIMESTAMP,
	PRIMARY KEY (id)
);

create table schedule
(
	id SERIAL NOT NULL,
	job_id varchar(64),
	status varchar(64),
	creation_time timestamp default CURRENT_TIMESTAMP,
	update_time timestamp default CURRENT_TIMESTAMP,
	PRIMARY KEY (id)
);
