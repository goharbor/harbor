/*
Add new column creator_ref and creator_type for robot table to record the creator information of the robot
*/
ALTER TABLE robot ADD COLUMN IF NOT EXISTS creator_ref integer default 0;
ALTER TABLE robot ADD COLUMN IF NOT EXISTS creator_type varchar(255);

ALTER TABLE p2p_preheat_policy ADD COLUMN IF NOT EXISTS scope varchar(255);