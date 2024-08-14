/*
Add new column creator for robot table to add a new column to record the creator of the robot
*/
ALTER TABLE robot ADD COLUMN IF NOT EXISTS creator varchar(255);
UPDATE robot SET creator = 'unknown' WHERE creator IS NULL;
