CALL PROC_ADD_COLUMN_IF_NOT_EXISTS('schedule', 'revision', 'integer');
UPDATE schedule set revision = 0;