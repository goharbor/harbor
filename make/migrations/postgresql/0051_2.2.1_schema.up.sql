/*fixes #14358*/
UPDATE execution SET status='Success' WHERE status='Succeed';

CREATE INDEX IF NOT EXISTS task_execution_id_idx ON task (execution_id);