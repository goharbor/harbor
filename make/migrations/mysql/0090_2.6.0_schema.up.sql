/* Correct project_metadata.public value, should only be true or false, other invaild value will be rewrite to false */
UPDATE project_metadata SET value='false' WHERE name='public' AND value NOT IN('true', 'false');

/*
System Artifact Manager
Github proposal link : https://github.com/goharbor/community/pull/181
*/
 CREATE TABLE IF NOT EXISTS system_artifact (
        id SERIAL NOT NULL PRIMARY KEY,
        repository varchar(256) NOT NULL,
        digest varchar(255) NOT NULL DEFAULT '' ,
        size bigint NOT NULL DEFAULT 0 ,
        vendor varchar(255) NOT NULL DEFAULT '' ,
        type varchar(255) NOT NULL DEFAULT '' ,
        create_time timestamp default CURRENT_TIMESTAMP,
        extra_attrs text NOT NULL DEFAULT '' ,
        UNIQUE (repository, digest,  vendor)
);

CALL PROC_CREATE_INDEX_IF_NOT_EXISTS('artifact', 'repository_name', 'idx_artifact_repository_name');

CALL PROC_CREATE_INDEX_IF_NOT_EXISTS('execution', 'vendor_type', 'idx_execution_vendor_type_vendor_id');
CALL PROC_CREATE_INDEX_IF_NOT_EXISTS('execution', 'start_time', 'idx_execution_start_time');
CALL PROC_CREATE_INDEX_IF_NOT_EXISTS('audit_log', 'project_id, op_time', 'idx_audit_log_project_id_optime');
