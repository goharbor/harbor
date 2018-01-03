ALTER TABLE `timestamp_keys` ADD COLUMN `role` VARCHAR(255) NOT NULL, DROP KEY `gun`, ADD UNIQUE KEY `gun_role` (`gun`, `role`);

UPDATE `timestamp_keys` SET `role`="timestamp";
