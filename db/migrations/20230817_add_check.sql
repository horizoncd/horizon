-- check table
CREATE TABLE `tb_check`
(
    `id`            bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `resource_type` varchar(64)         NOT NULL DEFAULT '' COMMENT 'resource type',
    `resource_id`   bigint(20) unsigned NOT NULL COMMENT 'resource id',
    `created_at`    datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`    datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_ts`    bigint(20)                   DEFAULT '0' COMMENT 'deleted timestamp, 0 means not deleted',
    `created_by`    bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'creator',
    `updated_by`    bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'updater',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_resource_deleted` (`resource_type`, `resource_id`, `deleted_ts`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- check run table
CREATE TABLE `tb_checkrun`
(
  `id`              bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `name`            varchar(256)        NOT NULL DEFAULT '' COMMENT 'the name of check run',
  `status`          varchar(64)         NOT NULL DEFAULT '' COMMENT 'the status of check run',
  `pipeline_run_id` bigint(20) unsigned NOT NULL COMMENT 'pipeline run id',
  `check_id`        bigint(20) unsigned NOT NULL COMMENT 'check id',
  `message`         varchar(256)        NOT NULL DEFAULT '',
  `detail_url`      varchar(256)        NOT NULL DEFAULT '' COMMENT 'the detail url of check run',
  `created_at`      datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_ts`      bigint(20)                   DEFAULT '0' COMMENT 'deleted timestamp, 0 means not deleted',
  `created_by`      bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'creator',
  `updated_by`      bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'updater',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_pipeline_run_id_check_id_deleted` (`pipeline_run_id`, `check_id`, `deleted_ts`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- pr_msg table
CREATE TABLE `tb_pr_msg`
(
  `id`              bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `pipeline_run_id` bigint(20) unsigned NOT NULL COMMENT 'pipeline run id',
  `content`         text                NOT NULL COMMENT 'content of message',
  `created_at`      datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `created_by`      bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'creator',
  `updated_by`      bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'updater',
  `deleted_ts`      bigint(20)                   DEFAULT '0' COMMENT 'deleted timestamp, 0 means not deleted',
  PRIMARY KEY (`id`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;
