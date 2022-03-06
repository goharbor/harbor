ALTER TABLE schedule ADD COLUMN revision integer;
UPDATE schedule set revision = 0;