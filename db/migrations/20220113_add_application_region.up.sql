-- application region table
CREATE TABLE `application_region`
(
    `id`               bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `application_id`   bigint(20) unsigned NOT NULL COMMENT 'application id',
    `environment_name` varchar(128) NOT NULL DEFAULT '' COMMENT 'environment name',
    `region_name`      varchar(128) NOT NULL DEFAULT '' COMMENT 'region name',
    `created_at`       datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`       datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `created_by`       bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'creator',
    `updated_by`       bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'updater',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_application_environment` (`application_id`, `environment_name`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;
