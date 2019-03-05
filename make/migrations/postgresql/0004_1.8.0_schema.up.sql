/*add robot account table*/
CREATE TABLE robot (
 id SERIAL PRIMARY KEY NOT NULL,
 name varchar(255),
 description varchar(1024),
 project_id int,
 expiresat bigint,
 disabled boolean DEFAULT false NOT NULL,
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP,
 CONSTRAINT unique_robot UNIQUE (name, project_id)
);

CREATE TRIGGER robot_update_time_at_modtime BEFORE UPDATE ON robot FOR EACH ROW EXECUTE PROCEDURE update_update_time_at_column();

/*add master role*/
INSERT INTO role (role_code, name) VALUES ('DRWS', 'master');

/*delete replication jobs whose policy has been marked as "deleted"*/
DELETE FROM replication_job AS j
USING replication_policy AS p
WHERE j.policy_id = p.id AND p.deleted = TRUE;

/*delete replication policy which has been marked as "deleted"*/
DELETE FROM replication_policy AS p
WHERE p.deleted = TRUE;