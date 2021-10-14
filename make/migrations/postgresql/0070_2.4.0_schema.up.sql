/* cleanup deleted user project members */
DELETE FROM project_member pm WHERE pm.entity_type = 'u' AND EXISTS (SELECT NULL FROM harbor_user u WHERE pm.entity_id = u.user_id AND u.deleted = true );

ALTER TABLE replication_policy ADD COLUMN IF NOT EXISTS speed_kb int;

/* add version fields for lock free quota */
ALTER TABLE quota ADD COLUMN IF NOT EXISTS version bigint DEFAULT 0;
ALTER TABLE quota_usage ADD COLUMN IF NOT EXISTS version bigint DEFAULT 0;

/* convert Negligible to None for the severity of the vulnerability record */
UPDATE vulnerability_record SET severity='None' WHERE severity='Negligible';
