/* cleanup deleted user project members */
DELETE FROM project_member WHERE project_member.entity_type = 'u' AND EXISTS (SELECT NULL FROM harbor_user WHERE project_member.entity_id = harbor_user.user_id AND harbor_user.deleted = true );

CALL PROC_ADD_COLUMN_IF_NOT_EXISTS('replication_policy', 'speed_kb', 'int');

/* add version fields for lock free quota */
ALTER TABLE quota ADD COLUMN version bigint DEFAULT 0;
ALTER TABLE quota_usage ADD COLUMN version bigint DEFAULT 0;
CALL PROC_ADD_COLUMN_IF_NOT_EXISTS('quota', 'version', 'bigint DEFAULT 0');
CALL PROC_ADD_COLUMN_IF_NOT_EXISTS('quota_usage', 'version', 'bigint DEFAULT 0');

/* convert Negligible to None for the severity of the vulnerability record */
UPDATE vulnerability_record SET severity='None' WHERE severity='Negligible';
