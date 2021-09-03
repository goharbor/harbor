/* cleanup deleted user project members */
DELETE FROM project_member pm WHERE pm.entity_type = 'u' AND EXISTS (SELECT NULL FROM harbor_user u WHERE pm.entity_id = u.user_id AND u.deleted = true );

ALTER TABLE replication_policy ADD COLUMN IF NOT EXISTS speed_kb int;
