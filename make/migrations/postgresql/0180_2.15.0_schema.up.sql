/*
Initialize skip_audit_log_database configuration based on existing audit log usage - Only insert the configuration if it doesn't already exist
1. If tables exist and show evidence of previous usage
   set skip_audit_log_database to false
2. If tables exist but show no evidence of usage, don't create the configuration record
*/
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM properties WHERE k = 'skip_audit_log_database') THEN
        RETURN;
    END IF;

    IF (SELECT last_value FROM audit_log_id_seq) > 1
       OR (SELECT last_value FROM audit_log_ext_id_seq) > 1 THEN
        INSERT INTO properties (k, v) VALUES ('skip_audit_log_database', 'false');
    END IF;
END $$;

ALTER TABLE registry ADD COLUMN IF NOT EXISTS ca_certificate TEXT;

ALTER TABLE artifact_accessory ADD COLUMN IF NOT EXISTS source varchar(50) DEFAULT '';