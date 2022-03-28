CALL PROC_CREATE_INDEX_IF_NOT_EXISTS('artifact', 'push_time', 'idx_artifact_push_time');
CALL PROC_CREATE_INDEX_IF_NOT_EXISTS('tag', 'push_time', 'idx_tag_push_time');
CALL PROC_CREATE_INDEX_IF_NOT_EXISTS('tag', 'artifact_id', 'idx_tag_artifact_id');
CALL PROC_CREATE_INDEX_IF_NOT_EXISTS('artifact_reference', 'child_id', 'idx_artifact_reference_child_id');