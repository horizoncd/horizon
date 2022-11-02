CREATE TABLE `tb_registry` (
  `id`                       bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `name`                     varchar(128) NOT NULL DEFAULT '' COMMENT 'name of the harbor registry',
  `server`                   varchar(256) NOT NULL DEFAULT '' COMMENT 'harbor server address',
  `token`                    varchar(512) NOT NULL DEFAULT '' COMMENT 'harbor server token',
  `path`                     varchar(256) NOT NULL DEFAULT '' COMMENT 'path of image',
  `insecure_skip_tls_verify` tinyint(1) NOT NULL DEFAULT false COMMENT 'skip tls verify',
  `kind`                     varchar(256) NOT NULL DEFAULT 'harbor' COMMENT 'which kind registry it is',
  `created_at`               datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`               datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_ts`               bigint(20) DEFAULT '0' COMMENT 'deleted timestamp, 0 means not deleted',
  `created_by`               bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'creator',
  `updated_by`               bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'updater',
  PRIMARY KEY (`id`)
) ENGINE = InnoDB 
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

INSERT INTO tb_registry (`id`, `name`, `server`, `token`, `insecure_skip_tls_verify`)
(SELECT `id`, `name`, `server`, `token`, true from tb_harbor);

ALTER TABLE tb_region add column `registry_id` bigint(20) unsigned NOT NULL COMMENT 'registry id';
UPDATE tb_region set `registry_id` = `harbor_id`;