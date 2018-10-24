ALTER TABLE properties ALTER COLUMN v TYPE varchar(1024);
DELETE FROM properties where k='scan_all_policy';
