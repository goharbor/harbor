DROP TABLE IF EXISTS harbor_resource_label;

CREATE INDEX IF NOT EXISTS idx_artifact_accessory_subject_artifact_id ON artifact_accessory (subject_artifact_id);