/*
Fix for issue #23772: Add indexed columns for sbom_digest to avoid
sequential scans caused by report::jsonb @> containment predicates
during artifact/tag retention.

For sbom_report: extract "sbom_digest" from the JSON report column.
For scan_report: extract "sbom_digest" from the JSON report column.

Both tables used: DELETE ... WHERE report::jsonb @> '{"sbom_digest":"..."}' 
which forced full sequential scans + jsonb casts on every deletion.
The new indexed columns allow simple equality lookups instead.
*/

/* --- sbom_report --- */
ALTER TABLE sbom_report ADD COLUMN IF NOT EXISTS sbom_digest VARCHAR(255);

/* Backfill existing rows from JSON report column */
UPDATE sbom_report
   SET sbom_digest = report::jsonb ->> 'sbom_digest'
 WHERE sbom_digest IS NULL
   AND report IS NOT NULL
   AND report::jsonb ->> 'sbom_digest' IS NOT NULL;

/* Btree index for the new column to support equality-based deletes */
CREATE INDEX IF NOT EXISTS idx_sbom_report_sbom_digest ON sbom_report (sbom_digest);

/* --- scan_report --- */
ALTER TABLE scan_report ADD COLUMN IF NOT EXISTS sbom_digest VARCHAR(255);

/* Backfill existing rows from JSON report column */
UPDATE scan_report
   SET sbom_digest = report::jsonb ->> 'sbom_digest'
 WHERE sbom_digest IS NULL
   AND report IS NOT NULL
   AND report::jsonb ->> 'sbom_digest' IS NOT NULL;

/* Btree index for the new column to support equality-based deletes */
CREATE INDEX IF NOT EXISTS idx_scan_report_sbom_digest ON scan_report (sbom_digest);
