-- cluster template schema tag table
CREATE TABLE `cluster_template_schema_tag`
(
    `id`         bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `cluster_id` bigint(20) unsigned NOT NULL COMMENT 'cluster id',
    `tag_key`        varchar(64) NOT NULL DEFAULT '' COMMENT 'key of tag',
    `tag_value`      varchar(1280) NOT NULL DEFAULT '' COMMENT 'value of tag',
    `created_at` datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `created_by` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'creator',
    `updated_by` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'updater',
    PRIMARY KEY (`id`),
    KEY          `idx_cluster_id` (`cluster_id`),
    KEY          `idx_key` (`tag_key`),
    UNIQUE KEY `idx_cluster_id_key` (`cluster_id`, `tag_key`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;
