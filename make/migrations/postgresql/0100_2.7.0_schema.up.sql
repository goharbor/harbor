ALTER TABLE replication_policy ADD COLUMN IF NOT EXISTS copy_by_chunk boolean;

CREATE TABLE IF NOT EXISTS job_queue_status (
        id SERIAL NOT NULL PRIMARY KEY,
        job_type varchar(256) NOT NULL,
        paused boolean NOT NULL DEFAULT false,
        update_time timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
        UNIQUE ("job_type")
);
