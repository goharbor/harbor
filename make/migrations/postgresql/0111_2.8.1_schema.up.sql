CREATE INDEX IF NOT EXISTS idx_task_extra_attrs_report_uuids ON task USING gin ((extra_attrs::jsonb->'report_uuids'));
