/* add capabilities in scanner_registration */
ALTER TABLE scanner_registration ADD COLUMN IF NOT EXISTS capabilities varchar(1024);