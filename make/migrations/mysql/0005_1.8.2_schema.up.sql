/*
Rename the duplicate names before adding "UNIQUE" constraint
*/
UPDATE user_group AS r
LEFT JOIN (
SELECT
	id,
IF
	( @gid = group_name, @idx := @idx + 1, 1 ) AS seq,
	@gid := group_name AS gid
FROM
	user_group,
	( SELECT @idx := 0, @gid = NULL ) t
ORDER BY
	group_name,
	id
	) v ON r.id = v.id
	SET r.group_name = (
/*
	truncate the name if it is too long after appending the sequence number
	*/
CASE

	WHEN ( length( group_name ) + length( concat( v.seq, '' ) ) + 1 ) > 256 THEN
	concat( substring( group_name, 1, ( 255- length( concat( v.seq, '' ) ) ) ), '_', v.seq ) ELSE concat( group_name, '_', v.seq )
END
	)
WHERE
v.seq > 1;

ALTER TABLE user_group ADD CONSTRAINT unique_group_name UNIQUE (group_name);


/*
Fix issue https://github.com/goharbor/harbor/issues/8526, delete the none scan_all schedule.
 */
UPDATE admin_job SET deleted='true' WHERE cron_str='{"type":"none"}';
