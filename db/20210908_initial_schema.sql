CREATE TABLE `group` (
    `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
    `name` varchar(128) NOT NULL DEFAULT '',
    `path` varchar(256) NOT NULL DEFAULT '',
    `description` varchar(256) DEFAULT NULL,
    `visibility_level` tinyint(1) NOT NULL,
    `parent_id` int(11) NOT NULL,
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` datetime DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_deleted_at_name` (`deleted_at`, `name`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;