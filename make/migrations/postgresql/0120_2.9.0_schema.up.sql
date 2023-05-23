CREATE INDEX IF NOT EXISTS idx_task_extra_attrs_report_uuids ON task USING gin ((extra_attrs::jsonb->'report_uuids'));

/* Set the vendor_id of IMAGE_SCAN to the artifact id instead of scanner id, which facilitates execution sweep */
UPDATE execution SET vendor_id = (extra_attrs -> 'artifact' ->> 'id')::integer
WHERE jsonb_path_exists(extra_attrs::jsonb, '$.artifact.id')
AND vendor_id IN (SELECT id FROM scanner_registration)
AND vendor_type = 'IMAGE_SCAN';