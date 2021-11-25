-- group table
CREATE TABLE `group`
(
    `id`               bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `name`             varchar(128) NOT NULL DEFAULT '',
    `path`             varchar(32)  NOT NULL DEFAULT '',
    `description`      varchar(256)          DEFAULT NULL,
    `visibility_level` varchar(16)  NOT NULL COMMENT 'public or private',
    `parent_id`        bigint(20) NOT NULL DEFAULT 0 COMMENT 'ID of the parent group',
    `traversal_ids`    varchar(32)  NOT NULL DEFAULT '' COMMENT 'ID path from the root, like 1,2,3',
    `created_at`       datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`       datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`       datetime              DEFAULT NULL,
    `created_by`       bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'creator',
    `updated_by`       bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'updater',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_parent_id_name_delete_at` (`parent_id`, `name`, `deleted_at`),
    UNIQUE KEY `idx_parent_id_path_delete_at` (`parent_id`, `path`, `deleted_at`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- user table
CREATE TABLE `user`
(
    `id`         bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `name`       varchar(64) NOT NULL DEFAULT '',
    `full_name`  varchar(128)         DEFAULT '',
    `email`      varchar(64) NOT NULL DEFAULT '',
    `phone`      varchar(32)          DEFAULT NULL COMMENT '',
    `oidc_id`    varchar(64) NOT NULL COMMENT 'oidc id, which is a unique key in oidc system.',
    `oidc_type`  varchar(64) NOT NULL COMMENT 'oidc type, such as google, github, gitlab etc.',
    `admin`      tinyint(1) NOT NULL COMMENT 'is system admin，0-false，1-true',
    `created_at` datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` datetime             DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY          `idx_deleted_at` (`deleted_at`),
    UNIQUE KEY `idx_name` (`name`),
    UNIQUE KEY `idx_email`(`email`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- template table
CREATE TABLE `template`
(
    `id`          bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `name`        varchar(64) NOT NULL DEFAULT '' COMMENT 'the name of template',
    `description` varchar(256) NULL COMMENT 'the template description',
    `created_at`  datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`  datetime             DEFAULT NULL,
    `created_by`  bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'creator',
    `updated_by`  bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'updater',
    PRIMARY KEY (`id`),
    KEY           `idx_deleted_at` (`deleted_at`),
    UNIQUE KEY `idx_name` (`name`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- template release table
CREATE TABLE `template_release`
(
    `id`             bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `template_name`  varchar(64)  NOT NULL COMMENT 'the name of template',
    `name`           varchar(64)  NOT NULL DEFAULT '' COMMENT 'the name of template release',
    `description`    varchar(256) NOT NULL COMMENT 'description about this template release',
    `gitlab_project` varchar(256) NOT NULL COMMENT 'project ID or relative path in gitlab',
    `recommended`    tinyint(1) NOT NULL COMMENT 'is the most recommended template, 0-false, 1-true',
    `created_at`     datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`     datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`     datetime              DEFAULT NULL,
    `created_by`     bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'creator',
    `updated_by`     bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'updater',
    PRIMARY KEY (`id`),
    KEY              `idx_deleted_at` (`deleted_at`),
    UNIQUE KEY `idx_template_name_name` (`template_name`, `name`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- member table
CREATE TABLE `member`
(
    `id`            bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `resource_type` varchar(64) NOT NULL COMMENT 'group\application\cluster',
    `resource_id`   bigint(20) unsigned NOT NULL COMMENT 'resource id',
    `role`          varchar(64) NOT NULL COMMENT 'binding role name',
    `member_type`   tinyint(1) COMMENT '0-USER, 1-group',
    `membername_id` bigint(20) unsigned NOT NULL COMMENT 'UserID or GroupID',
    `granted_by`    bigint(20) unsigned NOT NULL COMMENT 'who grant the role',
    `created_by`    bigint(20) unsigned NOT NULL COMMENT 'who create the role',
    `created_at`    datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`    datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`    datetime             DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY             `idx_deleted_at` (`deleted_at`),
    UNIQUE KEY `idx_resource_member` (`resource_type`, `resource_id`, `member_type`, `membername_id`, `deleted_at`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- application table
CREATE TABLE `application`
(
    `id`               bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `group_id`         bigint(20) unsigned NOT NULL COMMENT 'group id',
    `name`             varchar(64) NOT NULL DEFAULT '' COMMENT 'the name of application',
    `description`      varchar(256)         DEFAULT NULL COMMENT 'the description of application',
    `priority`         varchar(16) NOT NULL DEFAULT 'P3' COMMENT 'the priority of application',
    `git_url`          varchar(128)         DEFAULT NULL COMMENT 'git repo url',
    `git_subfolder`    varchar(128)         DEFAULT NULL COMMENT 'git repo subfolder',
    `git_branch`       varchar(128)         DEFAULT NULL COMMENT 'git default branch',
    `template`         varchar(64) NOT NULL COMMENT 'template name',
    `template_release` varchar(64) NOT NULL COMMENT 'template release',
    `created_at`       datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`       datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`       datetime             DEFAULT NULL,
    `created_by`       bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'creator',
    `updated_by`       bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'updater',
    PRIMARY KEY (`id`),
    KEY                `idx_deleted_at` (`deleted_at`),
    UNIQUE KEY `idx_name_deleted_at` (`name`, `deleted_at`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- k8s cluster table
CREATE TABLE `k8s_cluster`
(
    `id`             bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `name`           varchar(128) NOT NULL DEFAULT '' COMMENT 'k8s name',
    `server`         varchar(256) NOT NULL DEFAULT '' COMMENT 'k8s server',
    `certificate`    text         NOT NULL COMMENT 'k8s certificate',
    `ingress_domain` varchar(128)          DEFAULT NULL COMMENT 'k8s ingress address',
    `created_at`     datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`     datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`     datetime              DEFAULT NULL,
    `created_by`     bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'creator',
    `updated_by`     bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'updater',
    PRIMARY KEY (`id`),
    KEY              `idx_deleted_at` (`deleted_at`),
    UNIQUE KEY `idx_server`(`server`, `deleted_at`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- harbor table
CREATE TABLE `harbor`
(
    `id`                bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `server`            varchar(256) NOT NULL DEFAULT '' COMMENT 'harbor server address',
    `token`             varchar(512) NOT NULL COMMENT 'harbor server token',
    `preheat_policy_id` int(2) COMMENT 'p2p preheat policy id',
    `created_at`        datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`        datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`        datetime              DEFAULT NULL,
    `created_by`        bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'creator',
    `updated_by`        bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'updater',
    PRIMARY KEY (`id`),
    KEY                 `idx_deleted_at` (`deleted_at`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- environment table
CREATE TABLE `environment`
(
    `id`           bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `name`         varchar(128) NOT NULL DEFAULT '' COMMENT 'env name',
    `display_name` varchar(128) NOT NULL DEFAULT '' COMMENT 'display name',
    `created_at`   datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`   datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`   datetime              DEFAULT NULL,
    `created_by`   bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'creator',
    `updated_by`   bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'updater',
    PRIMARY KEY (`id`),
    KEY            `idx_deleted_at` (`deleted_at`),
    UNIQUE KEY `idx_name` (`name`, `deleted_at`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- region table
CREATE TABLE `region`
(
    `id`             bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `name`           varchar(128) NOT NULL DEFAULT '' COMMENT 'region name',
    `display_name`   varchar(128) NOT NULL DEFAULT '' COMMENT 'region display name',
    `k8s_cluster_id` bigint(20) unsigned NOT NULL COMMENT 'k8s cluster id',
    `harbor_id`      bigint(20) unsigned NOT NULL COMMENT 'harbor id',
    `created_at`     datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`     datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`     datetime              DEFAULT NULL,
    `created_by`     bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'creator',
    `updated_by`     bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'updater',
    PRIMARY KEY (`id`),
    KEY              `idx_deleted_at` (`deleted_at`),
    UNIQUE KEY `idx_name` (`name`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- environment_region table
CREATE TABLE `environment_region`
(
    `id`               bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `environment_name` varchar(128) NOT NULL DEFAULT '' COMMENT 'environment name',
    `region_name`      varchar(128) NOT NULL DEFAULT '' COMMENT 'region name',
    `disabled`         tinyint(1) NOT NULL DEFAULT 0 COMMENT 'is disabled，0-false，1-true',
    `created_at`       datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`       datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`       datetime              DEFAULT NULL,
    `created_by`       bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'creator',
    `updated_by`       bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'updater',
    PRIMARY KEY (`id`),
    KEY                `idx_deleted_at` (`deleted_at`),
    UNIQUE KEY `idx_env_region` (`environment_name`, `region_name`, `deleted_at`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- cluster table
CREATE TABLE `cluster`
(
    `id`                    bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `application_id`        bigint(20) unsigned NOT NULL COMMENT 'application id',
    `name`                  varchar(64)  NOT NULL DEFAULT '' COMMENT 'the name of cluster',
    `description`           varchar(256)          DEFAULT NULL COMMENT 'the description of cluster',
    `git_url`               varchar(128)          DEFAULT NULL COMMENT 'git repo url',
    `git_subfolder`         varchar(128)          DEFAULT NULL COMMENT 'git repo subfolder',
    `git_branch`            varchar(128)          DEFAULT NULL COMMENT 'git branch',
    `template`              varchar(64)  NOT NULL COMMENT 'template name',
    `template_release`      varchar(64)  NOT NULL COMMENT 'template release',
    `environment_region_id` varchar(128) NOT NULL DEFAULT '' COMMENT 'env',
    `status`                varchar(64)  NOT NULL DEFAULT '' COMMENT 'the status of cluster',
    `created_at`            datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`            datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`            datetime              DEFAULT NULL,
    `created_by`            bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'creator',
    `updated_by`            bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'updater',
    PRIMARY KEY (`id`),
    KEY                     `idx_deleted_at` (`deleted_at`),
    KEY                     `idx_application_id` (`application_id`),
    UNIQUE KEY `idx_name_deleted_at` (`name`, `deleted_at`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- cluster tag table
CREATE TABLE `cluster_tag`
(
    `id`         bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `cluster_id` bigint(20) unsigned NOT NULL COMMENT 'cluster id',
    `tag_key`    varchar(64)   NOT NULL DEFAULT '' COMMENT 'key of tag',
    `tag_value`  varchar(1280) NOT NULL DEFAULT '' COMMENT 'value of tag',
    `created_at` datetime      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `created_by` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'creator',
    `updated_by` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'updater',
    PRIMARY KEY (`id`),
    KEY          `idx_cluster_id` (`cluster_id`),
    KEY          `idx_key` (`tag_key`),
    UNIQUE KEY `idx_cluster_id_key` (`cluster_id`, `tag_key`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

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


-- pipelinerun table
CREATE TABLE `pipelinerun`
(
    `id`                 bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `cluster_id`         bigint(20) unsigned NOT NULL COMMENT 'cluster id',
    `action`             varchar(64)  NOT NULL COMMENT 'action',
    `status`             varchar(64)  NOT NULL DEFAULT '' COMMENT 'the pipelinerun status',
    `title`              varchar(256) NOT NULL DEFAULT '' COMMENT 'the title of pipelinerun',
    `description`        varchar(2048)         DEFAULT NULL COMMENT 'the description of pipelinerun',
    `git_url`            varchar(128)          DEFAULT NULL COMMENT 'git repo url',
    `git_branch`         varchar(128)          DEFAULT NULL COMMENT 'the branch to build of this pipelinerun',
    `git_commit`         varchar(128)          DEFAULT NULL COMMENT 'the commit to build of this pipelinerun',
    `image_url`          varchar(256)          DEFAULT NULL COMMENT 'image url',
    `last_config_commit` varchar(128)          DEFAULT NULL COMMENT 'the last commit of cluster config',
    `config_commit`      varchar(128)          DEFAULT NULL COMMENT 'the new commit of cluster config',
    `s3_bucket`          varchar(128) NOT NULL DEFAULT '' COMMENT 's3 bucket to storage this pipelinerun log',
    `log_object`         varchar(258) NOT NULL DEFAULT '' COMMENT 's3 object for log',
    `pr_object`          varchar(258) NOT NULL DEFAULT '' COMMENT 's3 object for pipelinerun',
    `started_at`         datetime              DEFAULT NULL COMMENT 'start time of this pipelinerun',
    `finished_at`        datetime              DEFAULT NULL COMMENT 'finish time of this pipelinerun',
    `rollback_from`      bigint(20) unsigned NULL COMMENT 'the pipelinerun id that this pipelinerun rollback from',
    `created_at`         datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`         datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `created_by`         bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'creator',
    PRIMARY KEY (`id`),
    KEY                  `idx_cluster_action` (`cluster_id`, `action`),
    KEY                  `idx_cluster_config_commit` (`cluster_id`, `config_commit`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;
