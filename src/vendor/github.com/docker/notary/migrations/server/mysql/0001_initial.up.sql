CREATE TABLE `timestamp_keys` (
	  `id` int(11) NOT NULL AUTO_INCREMENT,
	  `created_at` timestamp NULL DEFAULT NULL,
	  `updated_at` timestamp NULL DEFAULT NULL,
	  `deleted_at` timestamp NULL DEFAULT NULL,
	  `gun` varchar(255) NOT NULL,
	  `cipher` varchar(50) NOT NULL,
	  `public` blob NOT NULL,
	  PRIMARY KEY (`id`),
	  UNIQUE KEY `gun` (`gun`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `tuf_files` (
	  `id` int(11) NOT NULL AUTO_INCREMENT,
	  `created_at` timestamp NULL DEFAULT NULL,
	  `updated_at` timestamp NULL DEFAULT NULL,
	  `deleted_at` timestamp NULL DEFAULT NULL,
	  `gun` varchar(255) NOT NULL,
	  `role` varchar(255) NOT NULL,
	  `version` int(11) NOT NULL,
	  `data` longblob NOT NULL,
	  PRIMARY KEY (`id`),
	  UNIQUE KEY `gun` (`gun`,`role`,`version`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
