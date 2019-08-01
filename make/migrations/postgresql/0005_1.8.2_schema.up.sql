/*
Rename the duplicate names before adding "UNIQUE" constraint
*/
DO $$
BEGIN
    WHILE EXISTS (SELECT count(*) FROM user_group GROUP BY group_name HAVING count(*) > 1) LOOP
        UPDATE user_group AS r
        SET group_name = (
            /*
            truncate the name if it is too long after appending the sequence number
            */
            CASE WHEN (length(group_name)+length(v.seq::text)+1) > 256
            THEN
                substring(group_name from 1 for (255-length(v.seq::text))) || '_' || v.seq
            ELSE
                group_name || '_' || v.seq
            END
        )
        FROM (SELECT id, row_number() OVER (PARTITION BY group_name ORDER BY id) AS seq FROM user_group) AS v
        WHERE r.id = v.id AND v.seq > 1;
    END LOOP;
END $$;

ALTER TABLE user_group ADD CONSTRAINT unique_group_name UNIQUE (group_name);


/*
Fix issue https://github.com/goharbor/harbor/issues/8526, delete the none scan_all schedule.
 */
UPDATE admin_job SET deleted='true' WHERE cron_str='{"type":"none"}';
