/*
Convert setup_timestamp in p2p_preheat_instance, status_revision in task, and revision in schedule to bigint to avoid y2k38 overflow, issue #23711.
*/
ALTER TABLE p2p_preheat_instance ALTER COLUMN setup_timestamp TYPE bigint;
ALTER TABLE task ALTER COLUMN status_revision TYPE bigint;
ALTER TABLE schedule ALTER COLUMN revision TYPE bigint;
