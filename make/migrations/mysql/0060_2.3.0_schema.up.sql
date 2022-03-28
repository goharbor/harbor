CALL PROC_ADD_COLUMN_IF_NOT_EXISTS('replication_policy', 'dest_namespace_replace_count', 'int');
UPDATE replication_policy SET dest_namespace_replace_count=-1 WHERE dest_namespace IS NOT NULL;

CALL PROC_CREATE_INDEX_IF_NOT_EXISTS('artifact', 'push_time', 'idx_artifact_push_time');
CALL PROC_CREATE_INDEX_IF_NOT_EXISTS('tag', 'push_time', 'idx_tag_push_time');
CALL PROC_CREATE_INDEX_IF_NOT_EXISTS('tag', 'artifact_id', 'idx_tag_artifact_id');
CALL PROC_CREATE_INDEX_IF_NOT_EXISTS('artifact_reference', 'child_id', 'idx_artifact_reference_child_id');
CALL PROC_CREATE_INDEX_IF_NOT_EXISTS('audit_log', 'op_time', 'idx_audit_log_op_time');