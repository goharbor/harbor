ALTER TABLE artifact_accessory ADD COLUMN IF NOT EXISTS source varchar(50) DEFAULT 'local';

/* Track when an artifact row last changed (e.g. tag attached/detached). See #23149. */
ALTER TABLE artifact ADD COLUMN IF NOT EXISTS update_time timestamp DEFAULT CURRENT_TIMESTAMP;
UPDATE artifact SET update_time = COALESCE(push_time, CURRENT_TIMESTAMP) WHERE update_time IS NULL;
