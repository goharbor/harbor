insert into replication_target (name, url, username, password) values ('test', '192.168.0.2:5000', 'testuser', 'passw0rd');
insert into replication_policy (name, project_id, target_id, enabled, start_time) value ('test_policy', 1, 1, 1, NOW());
