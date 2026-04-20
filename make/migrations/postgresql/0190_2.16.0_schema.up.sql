ALTER TABLE artifact_accessory ADD COLUMN IF NOT EXISTS source varchar(50) DEFAULT 'local';
