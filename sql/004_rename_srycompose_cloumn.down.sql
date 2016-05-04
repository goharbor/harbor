use registry;

ALTER TABLE repository CHANGE COLUMN docker_compose  srycompose text;
ALTER TABLE repository DROP COLUMN marathon_config;

