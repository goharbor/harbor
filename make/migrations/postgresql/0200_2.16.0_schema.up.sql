ALTER TABLE sbom_report 
ADD COLUMN IF NOT EXISTS sbom_digest text 
GENERATED ALWAYS AS (report::jsonb ->> 'sbom_digest') STORED;

CREATE INDEX IF NOT EXISTS idx_sbom_report_sbom_digest 
ON sbom_report (mime_type, sbom_digest);

-- Do the exact same thing for the scan_report table too!
ALTER TABLE scan_report 
ADD COLUMN IF NOT EXISTS sbom_digest text 
GENERATED ALWAYS AS (report::jsonb ->> 'sbom_digest') STORED;

CREATE INDEX IF NOT EXISTS idx_scan_report_sbom_digest 
ON scan_report (mime_type, sbom_digest);
