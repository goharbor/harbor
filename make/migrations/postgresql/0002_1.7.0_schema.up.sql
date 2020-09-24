ALTER TABLE properties ALTER COLUMN v TYPE varchar(1024);
DELETE FROM properties where k='scan_all_policy';

create table job_log (
 log_id SERIAL NOT NULL,
 job_uuid varchar (64) NOT NULL,
 creation_time timestamp default CURRENT_TIMESTAMP,
 content text,
 primary key (log_id)
);

CREATE UNIQUE INDEX job_log_uuid ON job_log (job_uuid);

/*
Rename the duplicate names before adding "UNIQUE" constraint
*/
DO $$ 
BEGIN
    WHILE EXISTS (SELECT count(*) FROM replication_policy GROUP BY name HAVING count(*) > 1) LOOP
        UPDATE replication_policy AS r
        SET name = (
            /*
            truncate the name if it is too long after appending the sequence number
            */
            CASE WHEN (length(name)+length(v.seq::text)+1) > 256 
            THEN
                substring(name from 1 for (255-length(v.seq::text))) || '_' || v.seq
            ELSE
                name || '_' || v.seq
            END
        )
        FROM (SELECT id, row_number() OVER (PARTITION BY name ORDER BY id) AS seq FROM replication_policy) AS v
        WHERE r.id = v.id AND v.seq > 1;
    END LOOP;
END $$;

/*
Rename the duplicate names before adding "UNIQUE" constraint
*/
DO $$ 
BEGIN
    WHILE EXISTS (SELECT count(*) FROM replication_target GROUP BY name HAVING count(*) > 1) LOOP
        UPDATE replication_target AS t
        SET name = (
            CASE WHEN (length(name)+length(v.seq::text)+1) > 64 
            THEN
                substring(name from 1 for (63-length(v.seq::text))) || '_' || v.seq
            ELSE
                name || '_' || v.seq
            END
        )
        FROM (SELECT id, row_number() OVER (PARTITION BY name ORDER BY id) AS seq FROM replication_target) AS v
        WHERE t.id = v.id AND v.seq > 1;
    END LOOP;
END $$;

ALTER TABLE replication_policy ADD CONSTRAINT unique_policy_name UNIQUE (name);
ALTER TABLE replication_target ADD CONSTRAINT unique_target_name UNIQUE (name);
