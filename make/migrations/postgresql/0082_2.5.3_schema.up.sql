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