/* Remove old version scan reports of trivy */
DELETE FROM scan_report WHERE mime_type='application/vnd.scanner.adapter.vuln.report.harbor+json; version=1.0' AND registration_uuid IN (SELECT uuid FROM scanner_registration WHERE name='Trivy' AND immutable='true');

/* Change vulnerability_record.id and report_vulnerability_record.id to BIGINT */
ALTER TABLE vulnerability_record ALTER COLUMN id TYPE BIGINT;
ALTER SEQUENCE vulnerability_record_id_seq AS BIGINT;
ALTER TABLE report_vulnerability_record ALTER COLUMN id TYPE BIGINT;
ALTER SEQUENCE report_vulnerability_record_id_seq AS BIGINT;
