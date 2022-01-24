/* Remove old version scan reports of trivy */
DELETE FROM scan_report WHERE mime_type='application/vnd.scanner.adapter.vuln.report.harbor+json; version=1.0' AND registration_uuid IN (SELECT uuid FROM scanner_registration WHERE name='Trivy' AND immutable='true');
