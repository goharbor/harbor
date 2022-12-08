ALTER TABLE replication_policy ADD COLUMN IF NOT EXISTS copy_by_chunk boolean;

CREATE TABLE IF NOT EXISTS job_queue_status (
        id SERIAL NOT NULL PRIMARY KEY,
        job_type varchar(256) NOT NULL,
        paused boolean NOT NULL DEFAULT false,
        update_time timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
        UNIQUE ("job_type")
);

/* remove the redundant data from table artifact_blob */
delete from artifact_blob afb where not exists (select digest from blob b where b.digest = afb.digest_af);