ALTER TABLE schedule ADD COLUMN IF NOT EXISTS revision integer;
UPDATE schedule set revision = 0;