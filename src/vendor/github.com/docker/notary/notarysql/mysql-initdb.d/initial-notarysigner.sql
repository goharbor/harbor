CREATE DATABASE IF NOT EXISTS `notarysigner`;

CREATE USER "signer"@"%" IDENTIFIED BY "";

GRANT
	ALL PRIVILEGES ON `notarysigner`.* 
	TO "signer"@"%";
