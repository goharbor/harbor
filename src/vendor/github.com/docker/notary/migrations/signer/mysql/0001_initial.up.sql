CREATE TABLE `private_keys` (
	  `id` int(11) NOT NULL AUTO_INCREMENT,
	  `created_at` timestamp NULL DEFAULT NULL,
	  `updated_at` timestamp NULL DEFAULT NULL,
	  `deleted_at` timestamp NULL DEFAULT NULL,
	  `key_id` varchar(255) NOT NULL,
	  `encryption_alg` varchar(255) NOT NULL,
	  `keywrap_alg` varchar(255) NOT NULL,
	  `algorithm` varchar(50) NOT NULL,
	  `passphrase_alias` varchar(50) NOT NULL,
	  `public` blob NOT NULL,
	  `private` blob NOT NULL,
	  PRIMARY KEY (`id`),
	  UNIQUE KEY `key_id` (`key_id`),
	  UNIQUE KEY `key_id_2` (`key_id`,`algorithm`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
