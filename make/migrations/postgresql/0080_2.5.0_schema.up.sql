/* Fix retention_policy create_time, update_time for pgx driver */
ALTER TABLE retention_policy ALTER COLUMN create_time TYPE TIMESTAMP WITHOUT TIME ZONE USING (current_date + create_time);
ALTER TABLE retention_policy ALTER COLUMN update_time TYPE TIMESTAMP WITHOUT TIME ZONE USING (current_date + update_time);

/* create table of accessory */
CREATE TABLE IF NOT EXISTS artifact_accessory (
    id SERIAL PRIMARY KEY NOT NULL,
    /*
       the artifact id of the accessory itself.
    */
    artifact_id bigint,
    /*
     the subject artifact id of the accessory.
    */
    subject_artifact_id bigint,
    /*
     the type of the accessory, like signature.cosign.
    */
    type varchar(256),
    size bigint,
    digest varchar(1024),
    creation_time timestamp default CURRENT_TIMESTAMP,
    FOREIGN KEY (artifact_id) REFERENCES artifact(id),
    FOREIGN KEY (subject_artifact_id) REFERENCES artifact(id),
    CONSTRAINT unique_artifact_accessory UNIQUE (artifact_id, subject_artifact_id)
);

/* Change vulnerability_record.id and report_vulnerability_record.id to BIGINT */
ALTER TABLE vulnerability_record ALTER COLUMN id TYPE BIGINT;
ALTER SEQUENCE vulnerability_record_id_seq AS BIGINT;
ALTER TABLE report_vulnerability_record ALTER COLUMN id TYPE BIGINT;
ALTER SEQUENCE report_vulnerability_record_id_seq AS BIGINT;

CREATE INDEX IF NOT EXISTS idx_task_job_id ON task (job_id);


/*
System Artifact Manager
Github proposal link : https://github.com/goharbor/community/pull/181
*/

 CREATE TABLE IF NOT EXISTS "system_artifact" (
        "id" serial NOT NULL PRIMARY KEY,
        "repository" text NOT NULL DEFAULT '' ,
        "digest" text NOT NULL DEFAULT '' ,
        "size" bigint NOT NULL DEFAULT 0 ,
        "vendor" text NOT NULL DEFAULT '' ,
        "type" text NOT NULL DEFAULT '' ,
        "create_time" timestamp with time zone NOT NULL,
        "extra_attrs" text NOT NULL DEFAULT '' ,
        UNIQUE ("repository", "digest")
);
