use registry;
insert into replication_target (name, url, username, password) values ('test', 'http://10.117.171.31', 'admin', 'Harbor12345');
insert into replication_policy (name, project_id, target_id, enabled, start_time) value ('test_policy', 1, 1, 1, NOW());
insert into replication_job (status, policy_id, repository, operation) value ('running', 1, 'library/whatever', 'transfer')
