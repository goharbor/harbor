ALTER TABLE `tuf_files` ADD COLUMN `sha256` CHAR(64) DEFAULT NULL, ADD INDEX `sha256` (`sha256`);

-- SHA2 function takes the column name or a string as the first parameter, and the 
-- hash size as the second argument. It returns a hex string.
UPDATE `tuf_files` SET `sha256` = SHA2(`data`, 256);
