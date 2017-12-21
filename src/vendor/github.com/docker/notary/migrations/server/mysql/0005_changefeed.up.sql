CREATE TABLE `change_category` (
    `category` VARCHAR(20) NOT NULL,
    PRIMARY KEY (`category`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

INSERT INTO `change_category` VALUES ("update"), ("deletion");

CREATE TABLE `changefeed` (
    `id` int(11) NOT NULL AUTO_INCREMENT,
    `created_at` timestamp DEFAULT CURRENT_TIMESTAMP,
    `gun` varchar(255) NOT NULL,
    `version` int(11) NOT NULL,
    `sha256` CHAR(64) DEFAULT NULL,
    `category` VARCHAR(20) NOT NULL DEFAULT "update",
    PRIMARY KEY (`id`),
    FOREIGN KEY (`category`) REFERENCES `change_category` (`category`),
    INDEX `idx_changefeed_gun` (`gun`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

INSERT INTO `changefeed` (
        `created_at`,
        `gun`,
        `version`,
        `sha256` 
    ) (SELECT
        `created_at`,
        `gun`,
        `version`,
        `sha256`
    FROM
        `tuf_files`
    WHERE
        `role` = "timestamp"
    ORDER BY
        `created_at` ASC
);
