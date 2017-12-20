CREATE DATABASE IF NOT EXISTS `notaryserver`;

CREATE USER "server"@"%" IDENTIFIED BY "";

GRANT
	ALL PRIVILEGES ON `notaryserver`.* 
	TO "server"@"%";
