create table robot (
 id SERIAL PRIMARY KEY NOT NULL,
 name varchar(255),
 token varchar(4096),
 description varchar(1024),
 project_id int,
 disabled boolean DEFAULT false NOT NULL,
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP,
 CONSTRAINT unique_robot UNIQUE (name, project_id)
);

CREATE TRIGGER robot_update_time_at_modtime BEFORE UPDATE ON robot FOR EACH ROW EXECUTE PROCEDURE update_update_time_at_column();