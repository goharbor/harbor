/*
When upgrading from 2.0 to 2.1, the status and revision of retention schedule execution isn't migrated, correct them here.
As v2.2.0 isn't usable because of several serious bugs, we won't support upgrade from 2.1.4 to 2.2.0 anymore. After we add the
sql file here, users will get error when upgrading from 2.1.4 to 2.2.0 because of this sql file doesn't exist on 2.2.0
*/
UPDATE execution
SET revision=0, status=task.status
FROM task
WHERE execution.id=task.execution_id AND execution.vendor_type='SCHEDULER' AND execution.revision IS NULL;