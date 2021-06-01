CREATE INDEX IF NOT EXISTS idx_artifact_push_time ON artifact (push_time);
CREATE INDEX IF NOT EXISTS idx_tag_push_time ON tag (push_time);
CREATE INDEX IF NOT EXISTS idx_tag_artifact_id ON tag (artifact_id);
CREATE INDEX IF NOT EXISTS idx_artifact_reference_child_id ON artifact_reference (child_id);