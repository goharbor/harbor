/* Track when an artifact row last changed (e.g. tag attached/detached). See #23149. */
ALTER TABLE artifact ADD COLUMN IF NOT EXISTS update_time timestamp;
UPDATE artifact SET update_time = push_time WHERE update_time IS NULL;
ALTER TABLE artifact ALTER COLUMN update_time SET DEFAULT CURRENT_TIMESTAMP;
