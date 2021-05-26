ALTER TABLE replication_policy ADD COLUMN IF NOT EXISTS dest_namespace_replace_count int;
UPDATE replication_policy SET dest_namespace_replace_count=-1 WHERE dest_namespace IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_artifact_push_time ON artifact (push_time);
CREATE INDEX IF NOT EXISTS idx_tag_push_time ON tag (push_time);
CREATE INDEX IF NOT EXISTS idx_tag_artifact_id ON tag (artifact_id);
CREATE INDEX IF NOT EXISTS idx_artifact_reference_child_id ON artifact_reference (child_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_op_time ON audit_log (op_time);