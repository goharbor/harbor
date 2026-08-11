-- Prior to this release, prevent_vul=true also blocked pulling of unscanned/scan-failed
-- artifacts. That behavior is now controlled by the separate prevent_unscanned flag.
-- Backfill prevent_unscanned=true for projects that already had prevent_vul enabled, so
-- upgraded projects keep their previously enforced policy. Only projects without an
-- explicit prevent_unscanned value are touched; nothing set intentionally after upgrade
-- (including an explicit "false") is overwritten.
INSERT INTO project_metadata (project_id, name, value, creation_time, update_time)
SELECT pv.project_id, 'prevent_unscanned', 'true', NOW(), NOW()
FROM project_metadata pv
WHERE pv.name = 'prevent_vul'
  AND lower(pv.value) IN ('true', '1')
  AND NOT EXISTS (
    SELECT 1
    FROM project_metadata pu
    WHERE pu.project_id = pv.project_id
      AND pu.name = 'prevent_unscanned'
  );
