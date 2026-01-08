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

/* Fix for loop in 0120_2.9.0_schema.up.sql that never populated fixable_cnt.*/
DO
$$
    DECLARE
        report        RECORD;
        v             RECORD;
        fixable_count BIGINT;
    BEGIN
        FOR report IN SELECT uuid FROM scan_report WHERE fixable_cnt IS NULL
            LOOP
                fixable_count := 0;
                FOR v IN SELECT vr.fixed_version
                         FROM report_vulnerability_record rvr,
                              vulnerability_record vr
                         WHERE rvr.report_uuid = report.uuid
                           AND rvr.vuln_record_id = vr.id
                    LOOP
                        IF v.fixed_version IS NOT NULL AND v.fixed_version != '' THEN
                            fixable_count := fixable_count + 1;
                        END IF;
                    END LOOP;
                UPDATE scan_report
                SET fixable_cnt = fixable_count
                WHERE uuid = report.uuid;
            END LOOP;
    END
$$;
