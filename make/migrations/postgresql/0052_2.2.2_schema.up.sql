ALTER TABLE execution ADD COLUMN IF NOT EXISTS trigger_revision bigint;
UPDATE execution SET trigger_revision=id;
ALTER TABLE execution ADD CONSTRAINT unique_vendor_trigger_revision UNIQUE (vendor_type, vendor_id, trigger_revision);