CREATE DATABASE IF NOT EXISTS `notaryserver`;

CREATE USER "server"@"notary-server.%" IDENTIFIED BY "";

GRANT
	ALL PRIVILEGES ON `notaryserver`.* 
	TO "server"@"notary-server.%"
