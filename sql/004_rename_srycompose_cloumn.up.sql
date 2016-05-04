use registry;

ALTER TABLE repository CHANGE COLUMN srycompose docker_compose text;
ALTER TABLE repository ADD COLUMN marathon_config text;

