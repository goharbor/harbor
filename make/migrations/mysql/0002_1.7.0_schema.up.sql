ALTER TABLE properties MODIFY COLUMN v varchar(1024);
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
CREATE PROCEDURE PROC_UPDATE_REPLICATION_POLICY_NAME ( ) BEGIN
	WHILE
			EXISTS ( SELECT count( * ) FROM replication_policy GROUP BY NAME HAVING count( * ) > 1 ) DO
			UPDATE replication_policy AS t,
			(
			SELECT
				id,
				@idx :=
			IF
				( @gid = NAME, @idx + 1, 1 ) AS seq,
				@gid := NAME AS gid
			FROM
				replication_policy,
				( SELECT @idx := 0, @gid := NULL ) AS tt
			ORDER BY
				NAME,
				id
			) AS v
			SET NAME = (
            /*
            truncate the name if it is too long after appending the sequence number
            */
			CASE
					WHEN ( length( NAME ) + length( CONCAT( v.seq, '' ) ) + 1 ) > 256 THEN
					CONCAT( substring( NAME, 1, ( 255- length( CONCAT( v.seq, '' ) ) ) ), '_', v.seq ) ELSE CONCAT( NAME, '_', v.seq )
				END
				)
			WHERE
				t.id = v.id
				AND v.seq > 1;

		END WHILE;
END;

CALL PROC_UPDATE_REPLICATION_POLICY_NAME();

/*
Rename the duplicate names before adding "UNIQUE" constraint
*/
CREATE PROCEDURE PROC_UPDATE_REPLICATION_TARGET_NAME ( ) BEGIN
	WHILE
			EXISTS ( SELECT count( * ) FROM replication_target GROUP BY NAME HAVING count( * ) > 1 ) DO
			UPDATE replication_target AS t,
			(
			SELECT
				id,
				@idx :=
			IF
				( @gid = NAME, @idx + 1, 1 ) AS seq,
				@gid := NAME AS gid
			FROM
				replication_target,
				( SELECT @idx := 0, @gid := NULL ) AS tt
			ORDER BY
				NAME,
				id
			) AS v
			SET NAME = (
            /*
            truncate the name if it is too long after appending the sequence number
            */
			CASE

					WHEN ( length( NAME ) + length( CONCAT( v.seq, '' ) ) + 1 ) > 256 THEN
					CONCAT( substring( NAME, 1, ( 255- length( CONCAT( v.seq, '' ) ) ) ), '_', v.seq ) ELSE CONCAT( NAME, '_', v.seq )
				END
				)
			WHERE
				t.id = v.id
				AND v.seq > 1;

		END WHILE;
END;
CALL PROC_UPDATE_REPLICATION_TARGET_NAME ( );

ALTER TABLE replication_policy ADD CONSTRAINT unique_policy_name UNIQUE (name);
ALTER TABLE replication_target ADD CONSTRAINT unique_target_name UNIQUE (name);
