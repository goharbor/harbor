/*
Initialize skip_audit_log_database configuration based on existing audit log usage - Only insert the configuration if it doesn't already exist
1. If audit logs exist in the system (either in audit_log or audit_log_ext tables),
   set skip_audit_log_database to false
2. If no audit logs exist but the tables show evidence of previous usage
   set skip_audit_log_database to false
3. If tables exist but show no evidence of usage, don't create the configuration record

*/
DO $$
DECLARE
    audit_log_count INTEGER := 0;
    audit_log_ext_count INTEGER := 0;
    total_audit_logs INTEGER := 0;
    config_exists INTEGER := 0;
    audit_log_seq_value BIGINT := 0;
    audit_log_ext_seq_value BIGINT := 0;
    audit_log_table_used BOOLEAN := FALSE;
    audit_log_ext_table_used BOOLEAN := FALSE;
    should_set_config BOOLEAN := FALSE;
    skip_audit_value TEXT := 'false';
BEGIN
    SELECT COUNT(*) INTO config_exists
    FROM properties
    WHERE k = 'skip_audit_log_database';

    IF config_exists = 0 THEN
        SELECT COUNT(*) INTO audit_log_count FROM audit_log;
        SELECT last_value INTO audit_log_seq_value FROM audit_log_id_seq;
        audit_log_table_used := (audit_log_seq_value > 1);

        SELECT COUNT(*) INTO audit_log_ext_count FROM audit_log_ext;
        SELECT last_value INTO audit_log_ext_seq_value FROM audit_log_ext_id_seq;
        audit_log_ext_table_used := (audit_log_ext_seq_value > 1);

        total_audit_logs := audit_log_count + audit_log_ext_count;

        IF total_audit_logs > 0 THEN
            should_set_config := TRUE;
        ELSIF audit_log_table_used OR audit_log_ext_table_used THEN
            should_set_config := TRUE;
        END IF;

        IF should_set_config THEN
            INSERT INTO properties (k, v) VALUES ('skip_audit_log_database', skip_audit_value);
        END IF;
    END IF;
END $$;
