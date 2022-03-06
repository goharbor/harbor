/*fixes #14358*/
UPDATE execution SET status='Success' WHERE status='Succeed';

CREATE INDEX task_execution_id_idx ON task (execution_id);