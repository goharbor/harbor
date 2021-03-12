/*
When upgrading from 2.0 to 2.1, the status and revision of retention schedule execution isn't migrated, correct them here.
*/
UPDATE execution
SET revision=0, status=task.status
FROM task
WHERE execution.id=task.execution_id AND execution.vendor_type='SCHEDULER' AND execution.revision IS NULL;