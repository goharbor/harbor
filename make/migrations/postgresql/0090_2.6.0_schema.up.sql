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
        UNIQUE ("repository", "digest",  "vendor")
);

CREATE INDEX IF NOT EXISTS idx_artifact_repository_name ON artifact (repository_name);

CREATE INDEX IF NOT EXISTS idx_execution_vendor_type_vendor_id ON execution (vendor_type, vendor_id);
CREATE INDEX IF NOT EXISTS idx_execution_start_time ON execution(start_time);
CREATE INDEX IF NOT EXISTS idx_audit_log_project_id_optime ON audit_log (project_id, op_time);

/* repair execution status */
DO $$
DECLARE
    exec RECORD;
    status_group RECORD;
    status_count int;
    final_status varchar(32);
BEGIN
    /* iterate all executions */
    FOR exec IN SELECT * FROM execution WHERE status='Running'
    LOOP
        /* identify incorrect execution status, group tasks belong it by status */
        status_count = 0;
        final_status = '';
        FOR status_group IN SELECT status FROM task WHERE execution_id=exec.id GROUP BY status
        /* loop here to ensure all the tasks belong to the execution are success */
        LOOP
            status_count = status_count + 1;
            final_status = status_group.status;
        END LOOP;
        /* update status and end_time when the tasks are all
        success but itself status is not success */
        IF status_count=1 AND final_status='Success' THEN
            UPDATE execution SET status='Success', revision=revision+1 WHERE id=exec.id;
            UPDATE execution SET end_time=(SELECT MAX(end_time) FROM task WHERE execution_id=exec.id) WHERE id=exec.id;
        END IF;
    END LOOP;
END $$;

/* Add indexes to improve the performance of tag retention */
CREATE INDEX IF NOT EXISTS idx_artifact_blob_digest_blob ON artifact_blob (digest_blob);
CREATE INDEX IF NOT EXISTS idx_artifact_digest_project_id ON artifact (digest,project_id);