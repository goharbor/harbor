ALTER TABLE p2p_preheat_policy DROP COLUMN IF EXISTS scope;
ALTER TABLE p2p_preheat_policy ADD COLUMN IF NOT EXISTS extra_attrs text;

CREATE TABLE IF NOT EXISTS audit_log_ext
(
	id BIGSERIAL PRIMARY KEY NOT NULL,
	project_id BIGINT,
	operation VARCHAR(50) NULL,
	resource_type VARCHAR(255) NULL,
	resource VARCHAR(1024) NULL,
	username VARCHAR(255) NULL,
	op_desc VARCHAR(1024) NULL,
	op_result BOOLEAN DEFAULT true,
	payload TEXT NULL,
	source_ip VARCHAR(50) NULL,
	op_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- add index to the audit_log_ext table
CREATE INDEX IF NOT EXISTS idx_audit_log_ext_op_time ON audit_log_ext (op_time);
CREATE INDEX IF NOT EXISTS idx_audit_log_ext_project_id_optime ON audit_log_ext (project_id, op_time);
CREATE INDEX IF NOT EXISTS idx_audit_log_ext_project_id_resource_type ON audit_log_ext (project_id, resource_type);
CREATE INDEX IF NOT EXISTS idx_audit_log_ext_project_id_operation ON audit_log_ext (project_id, operation);
