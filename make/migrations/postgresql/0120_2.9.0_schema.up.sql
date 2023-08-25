CREATE INDEX IF NOT EXISTS idx_task_extra_attrs_report_uuids ON task USING gin ((extra_attrs::jsonb->'report_uuids'));

/* Set the vendor_id of IMAGE_SCAN to the artifact id instead of scanner id, which facilitates execution sweep */
UPDATE execution SET vendor_id = (extra_attrs -> 'artifact' ->> 'id')::integer
WHERE jsonb_path_exists(extra_attrs::jsonb, '$.artifact.id')
AND vendor_id IN (SELECT id FROM scanner_registration)
AND vendor_type = 'IMAGE_SCAN';

/* extract score from vendor attribute */
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM vulnerability_record WHERE cvss_score_v3 IS NOT NULL LIMIT 1) THEN
        UPDATE vulnerability_record
        SET cvss_score_v3 = (vendor_attributes->'CVSS'->'nvd'->>'V3Score')::double precision
        WHERE jsonb_path_exists(vendor_attributes::jsonb, '$.CVSS.nvd.V3Score');
    END IF;
END $$;

/* to improve the query of dangerousCVESQL it requires to query with vuln_record_id */
CREATE INDEX IF NOT EXISTS idx_report_vulnerability_record_vuln_record_id ON report_vulnerability_record (vuln_record_id);

CREATE INDEX IF NOT EXISTS idx_vulnerability_record_cvss_score_v3 ON vulnerability_record (cvss_score_v3);
CREATE INDEX IF NOT EXISTS idx_vulnerability_registration_uuid ON vulnerability_record (registration_uuid);
CREATE INDEX IF NOT EXISTS idx_vulnerability_record_cve_id ON vulnerability_record (cve_id);
CREATE INDEX IF NOT EXISTS idx_vulnerability_record_severity ON vulnerability_record (severity);
CREATE INDEX IF NOT EXISTS idx_vulnerability_record_package ON vulnerability_record (package);

/* add summary information in scan_report */
ALTER TABLE scan_report ADD COLUMN IF NOT EXISTS critical_cnt BIGINT;
ALTER TABLE scan_report ADD COLUMN IF NOT EXISTS high_cnt BIGINT;
ALTER TABLE scan_report ADD COLUMN IF NOT EXISTS medium_cnt BIGINT;
ALTER TABLE scan_report ADD COLUMN IF NOT EXISTS low_cnt BIGINT;
ALTER TABLE scan_report ADD COLUMN IF NOT EXISTS none_cnt BIGINT;
ALTER TABLE scan_report ADD COLUMN IF NOT EXISTS unknown_cnt BIGINT;
ALTER TABLE scan_report ADD COLUMN IF NOT EXISTS fixable_cnt BIGINT;

/* extract summary information for previous scan_report */
DO
$$
    DECLARE
        report RECORD;
        v RECORD;
        critical_count BIGINT;
        high_count BIGINT;
        none_count BIGINT;
        medium_count BIGINT;
        low_count BIGINT;
        unknown_count BIGINT;
        fixable_count BIGINT;
    BEGIN
        IF EXISTS (SELECT 1
                   FROM scan_report
                   WHERE critical_cnt IS NOT NULL
                     AND high_cnt IS NOT NULL
                     AND medium_cnt IS NOT NULL
                     AND low_cnt IS NOT NULL
                     AND unknown_cnt IS NOT NULL
                     AND fixable_cnt IS NOT NULL
                   LIMIT 1) THEN
            RETURN;
        END IF;

        FOR report IN SELECT uuid FROM scan_report
            LOOP
                critical_count := 0;
                high_count := 0;
                medium_count := 0;
                none_count := 0;
                low_count := 0;
                unknown_count := 0;
                FOR v IN SELECT vr.severity, vr.fixed_version
                         FROM report_vulnerability_record rvr,
                              vulnerability_record vr
                         WHERE rvr.report_uuid = report.uuid
                           AND rvr.vuln_record_id = vr.id
                    LOOP
                        IF v.severity = 'Critical' THEN
                            critical_count = critical_count + 1;
                        ELSIF v.severity = 'High' THEN
                            high_count = high_count + 1;
                        ELSIF v.severity = 'Medium' THEN
                            medium_count = medium_count + 1;
                        ELSIF v.severity = 'Low' THEN
                            low_count = low_count + 1;
                        ELSIF v.severity = 'None' THEN
                            none_count = none_count + 1;
                        ELSIF v.severity = 'Unknown' THEN
                            unknown_count = unknown_count + 1;
                        ELSIF v.fixed_version IS NOT NULL THEN
                            fixable_count = fixable_count + 1;
                        END IF;
                    END LOOP;
                UPDATE scan_report
                SET critical_cnt = critical_count,
                    high_cnt     = high_count,
                    medium_cnt   = medium_count,
                    low_cnt      = low_count,
                    unknown_cnt  = unknown_count
                WHERE uuid = report.uuid;
            END LOOP;
    END
$$;

/* Refactor the structure of replication schedule callback_func_param, convert the raw id to json object for extending */
/*       callback_func_param
    Old:         100
    New:  {"policy_id": 100}
*/
UPDATE schedule SET callback_func_param = json_build_object('policy_id', callback_func_param::int)::text
WHERE vendor_type='REPLICATION'
AND callback_func_param NOT LIKE '%policy_id%';