-- group table
CREATE TABLE `tb_group`
(
    `id`               bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `name`             varchar(128)        NOT NULL DEFAULT '',
    `path`             varchar(32)         NOT NULL DEFAULT '',
    `description`      varchar(256)                 DEFAULT NULL,
    `visibility_level` varchar(16)         NOT NULL COMMENT 'public or private',
    `parent_id`        bigint(20)          NOT NULL DEFAULT '0' COMMENT 'ID of the parent group',
    `traversal_ids`    varchar(32)         NOT NULL DEFAULT '' COMMENT 'ID path from the root, like 1,2,3',
    `region_selector`  varchar(512)        NOT NULL DEFAULT '' COMMENT 'used for filtering kubernetes',
    `created_at`       datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`       datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_ts`       bigint(20)                   DEFAULT '0' COMMENT 'deleted timestamp, 0 means not deleted',
    `created_by`       bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'creator',
    `updated_by`       bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'updater',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_parentId_name_deletedTs` (`parent_id`, `name`, `deleted_ts`),
    UNIQUE KEY `uk_parentId_path_deletedTs` (`parent_id`, `path`, `deleted_ts`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- user table
CREATE TABLE `tb_user`
(
    `id`         bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `name`       varchar(64)         NOT NULL DEFAULT '',
    `full_name`  varchar(128)                 DEFAULT '',
    `email`      varchar(64)         NOT NULL DEFAULT '',
    `phone`      varchar(32)                  DEFAULT NULL,
    `oidc_id`    varchar(64)         NOT NULL COMMENT 'oidc id, which is a unique index in oidc system.',
    `oidc_type`  varchar(64)         NOT NULL COMMENT 'oidc type, such as google, github, gitlab etc.',
    `admin`      tinyint(1)          NOT NULL COMMENT 'is system admin，0-false，1-true',
    `created_at` datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_ts` bigint(20)                   DEFAULT '0' COMMENT 'deleted timestamp, 0 means not deleted',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_name` (`name`),
    UNIQUE KEY `idx_email` (`email`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- template table
CREATE TABLE `tb_template`
(
    `id`          bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `name`        varchar(64)         NOT NULL DEFAULT '' COMMENT 'the name of template',
    `description` varchar(256)                 DEFAULT NULL COMMENT 'the template description',
    `repository`  varchar(256)        NOT NULL DEFAULT '',
    `group_id`    bigint(20) unsigned NOT NULL DEFAULT '0',
    `chart_name`  varchar(256)                 DEFAULT '',
    `only_admin`  tinyint(1)          NOT NULL DEFAULT '0',
    `created_at`  datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_ts`  bigint(20)                   DEFAULT '0' COMMENT 'deleted timestamp, 0 means not deleted',
    `created_by`  bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'creator',
    `updated_by`  bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'updater',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_name` (`name`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- template release table
CREATE TABLE `tb_template_release`
(
    `id`             bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `template_name`  varchar(64)         NOT NULL COMMENT 'the name of template',
    `name`           varchar(64)         NOT NULL DEFAULT '' COMMENT 'the name of template release',
    `description`    varchar(256)        NOT NULL COMMENT 'description about this template release',
    `gitlab_project` varchar(256)        NOT NULL COMMENT 'project ID or relative path in gitlab',
    `recommended`    tinyint(1)          NOT NULL COMMENT 'is the most recommended template, 0-false, 1-true',
    `template`       bigint(20) unsigned NOT NULL DEFAULT '0',
    `chart_name`     varchar(256)        NOT NULL DEFAULT '',
    `only_admin`     tinyint(1)          NOT NULL DEFAULT '0',
    `chart_version`  varchar(256)        NOT NULL DEFAULT '' COMMENT 'chart version on template repository',
    `sync_status`    varchar(64)         NOT NULL DEFAULT 'status_unknown' COMMENT 'shows sync status',
    `failed_reason`  varchar(2048)       NOT NULL DEFAULT '' COMMENT 'failed reason at last time',
    `commit_id`      varchar(256)        NOT NULL DEFAULT '' COMMENT 'commit id at last sync',
    `last_sync_at`   datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `created_at`     datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`     datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_ts`     bigint(20)                   DEFAULT '0' COMMENT 'deleted timestamp, 0 means not deleted',
    `created_by`     bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'creator',
    `updated_by`     bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'updater',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_template_name_name` (`template_name`, `name`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- member table
CREATE TABLE `tb_member`
(
    `id`            bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `resource_type` varchar(64)         NOT NULL COMMENT 'groupapplicationcluster',
    `resource_id`   bigint(20) unsigned NOT NULL COMMENT 'resource id',
    `role`          varchar(64)         NOT NULL COMMENT 'binding role name',
    `member_type`   tinyint(1)          NOT NULL DEFAULT '0' COMMENT '0-USER, 1-group',
    `membername_id` bigint(20) unsigned NOT NULL COMMENT 'UserID or GroupID',
    `granted_by`    bigint(20) unsigned NOT NULL COMMENT 'who grant the role',
    `created_by`    bigint(20) unsigned NOT NULL COMMENT 'who create the role',
    `created_at`    datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`    datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_ts`    bigint(20)          NOT NULL DEFAULT '0' COMMENT 'deleted timestamp, 0 means not deleted',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_resource_member_deleted` (`resource_type`, `resource_id`, `member_type`, `membername_id`,
                                             `deleted_ts`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- application table
CREATE TABLE `tb_application`
(
    `id`               bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `group_id`         bigint(20) unsigned NOT NULL COMMENT 'group id',
    `name`             varchar(64)         NOT NULL DEFAULT '' COMMENT 'the name of application',
    `description`      varchar(256)                 DEFAULT NULL COMMENT 'the description of application',
    `priority`         varchar(16)         NOT NULL DEFAULT 'P3' COMMENT 'the priority of application',
    `git_url`          varchar(128)                 DEFAULT NULL COMMENT 'git repo url',
    `git_subfolder`    varchar(128)                 DEFAULT NULL COMMENT 'git repo subfolder',
    `git_branch`       varchar(128)                 DEFAULT NULL COMMENT 'git default branch',
    `git_ref`          varchar(128)                 DEFAULT NULL,
    `git_ref_type`     varchar(64)                  DEFAULT NULL,
    `template`         varchar(64)         NOT NULL COMMENT 'template name',
    `template_release` varchar(64)         NOT NULL COMMENT 'template release',
    `created_at`       datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`       datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_ts`       bigint(20)                   DEFAULT '0' COMMENT 'deleted timestamp, 0 means not deleted',
    `created_by`       bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'creator',
    `updated_by`       bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'updater',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_name_deletedTs` (`name`, `deleted_ts`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- harbor table
CREATE TABLE `tb_harbor`
(
    `id`                bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `name`              varchar(128)        NOT NULL DEFAULT '' COMMENT 'name of the harbor registry',
    `server`            varchar(256)        NOT NULL DEFAULT '' COMMENT 'harbor server address',
    `token`             varchar(512)        NOT NULL COMMENT 'harbor server token',
    `preheat_policy_id` int(2)                       DEFAULT NULL COMMENT 'p2p preheat policy id',
    `created_at`        datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`        datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_ts`        bigint(20)                   DEFAULT '0' COMMENT 'deleted timestamp, 0 means not deleted',
    `created_by`        bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'creator',
    `updated_by`        bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'updater',
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- environment table
CREATE TABLE `tb_environment`
(
    `id`             bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `name`           varchar(128)        NOT NULL DEFAULT '' COMMENT 'env name',
    `display_name`   varchar(128)        NOT NULL DEFAULT '' COMMENT 'display name',
    `default_region` varchar(128)                 DEFAULT NULL COMMENT 'default region of the environment',
    `auto_free`      tinyint(1)          NOT NULL DEFAULT '0' COMMENT 'auto free configuration, 0 means disabled',
    `created_at`     datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`     datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_ts`     bigint(20)                   DEFAULT '0' COMMENT 'deleted timestamp, 0 means not deleted',
    `created_by`     bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'creator',
    `updated_by`     bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'updater',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_name_deletedTs` (`name`, `deleted_ts`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- region table
CREATE TABLE `tb_region`
(
    `id`             bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `name`           varchar(128)        NOT NULL DEFAULT '' COMMENT 'region name',
    `display_name`   varchar(128)        NOT NULL DEFAULT '' COMMENT 'region display name',
    `server`         varchar(256)                 DEFAULT NULL COMMENT 'k8s server url',
    `certificate`    text COMMENT 'k8s kube config',
    `ingress_domain` text COMMENT 'k8s ingress domain',
    `prometheus_url` varchar(128) COMMENT 'prometheus url',
    `harbor_id`      bigint(20) unsigned NOT NULL COMMENT 'harbor id',
    `disabled`       tinyint(1)          NOT NULL DEFAULT '0' COMMENT '0 means not disabled, 1 means disabled',
    `created_at`     datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`     datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_ts`     bigint(20)                   DEFAULT '0' COMMENT 'deleted timestamp, 0 means not deleted',
    `created_by`     bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'creator',
    `updated_by`     bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'updater',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_name` (`name`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- environment_region table
CREATE TABLE `tb_environment_region`
(
    `id`               bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `environment_name` varchar(128)        NOT NULL DEFAULT '' COMMENT 'environment name',
    `region_name`      varchar(128)        NOT NULL DEFAULT '' COMMENT 'region name',
    `is_default`       tinyint(1)          NOT NULL DEFAULT '0' COMMENT '0 means not default region, 1 means default region',
    `disabled`         tinyint(1)          NOT NULL DEFAULT '0' COMMENT 'is disabled，0-false，1-true',
    `created_at`       datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`       datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_ts`       bigint(20)                   DEFAULT '0' COMMENT 'deleted timestamp, 0 means not deleted',
    `created_by`       bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'creator',
    `updated_by`       bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'updater',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_env_region_deletedTs` (`environment_name`, `region_name`, `deleted_ts`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- cluster table
CREATE TABLE `tb_cluster`
(
    `id`               bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `application_id`   bigint(20) unsigned NOT NULL COMMENT 'application id',
    `name`             varchar(64)         NOT NULL DEFAULT '' COMMENT 'the name of cluster',
    `environment_name` varchar(128)        NOT NULL DEFAULT '',
    `region_name`      varchar(128)        NOT NULL DEFAULT '',
    `description`      varchar(256)                 DEFAULT NULL COMMENT 'the description of cluster',
    `git_url`          varchar(128)                 DEFAULT NULL COMMENT 'git repo url',
    `git_subfolder`    varchar(128)                 DEFAULT NULL COMMENT 'git repo subfolder',
    `git_branch`       varchar(128)                 DEFAULT NULL COMMENT 'git branch',
    `git_ref`          varchar(128)                 DEFAULT NULL,
    `git_ref_type`     varchar(64)                  DEFAULT NULL,
    `template`         varchar(64)         NOT NULL COMMENT 'template name',
    `template_release` varchar(64)         NOT NULL COMMENT 'template release',
    `status`           varchar(64)                  DEFAULT NULL,
    `expire_seconds`   bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'expiration seconds, 0 means permanent',
    `created_at`       datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`       datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_ts`       bigint(20)                   DEFAULT '0' COMMENT 'deleted timestamp, 0 means not deleted',
    `created_by`       bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'creator',
    `updated_by`       bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'updater',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_name_deletedTs` (`name`, `deleted_ts`),
    KEY `idx_application_id` (`application_id`),
    KEY `idx_deleted_ts` (`deleted_ts`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- tag table
CREATE TABLE `tb_tag`
(
    `id`            bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `resource_id`   bigint(20) unsigned NOT NULL COMMENT 'resource id',
    `resource_type` varchar(64)         NOT NULL DEFAULT '' COMMENT 'resource type',
    `tag_key`       varchar(64)         NOT NULL DEFAULT '' COMMENT 'key of tag',
    `tag_value`     varchar(1280)       NOT NULL DEFAULT '' COMMENT 'value of tag',
    `created_at`    datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`    datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `created_by`    bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'creator',
    `updated_by`    bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'updater',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_rType_cId_tKey` (`resource_type`, `resource_id`, `tag_key`),
    KEY `idx_cluster_id` (`resource_id`),
    KEY `idx_key` (`tag_key`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- cluster template schema tag table
CREATE TABLE `tb_cluster_template_schema_tag`
(
    `id`         bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `cluster_id` bigint(20) unsigned NOT NULL COMMENT 'cluster id',
    `tag_key`    varchar(64)         NOT NULL DEFAULT '' COMMENT 'key of tag',
    `tag_value`  varchar(1280)       NOT NULL DEFAULT '' COMMENT 'value of tag',
    `created_at` datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `created_by` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'creator',
    `updated_by` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'updater',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_cluster_id_key` (`cluster_id`, `tag_key`),
    KEY `idx_key` (`tag_key`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- pipelinerun table
CREATE TABLE `tb_pipelinerun`
(
    `id`                 bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `cluster_id`         bigint(20) unsigned NOT NULL COMMENT 'cluster id',
    `action`             varchar(64)         NOT NULL COMMENT 'action',
    `status`             varchar(64)         NOT NULL DEFAULT '' COMMENT 'the pipelinerun status',
    `title`              varchar(256)        NOT NULL DEFAULT '' COMMENT 'the title of pipelinerun',
    `description`        varchar(2048)                DEFAULT NULL COMMENT 'the description of pipelinerun',
    `git_url`            varchar(128)                 DEFAULT NULL COMMENT 'git repo url',
    `git_branch`         varchar(128)                 DEFAULT NULL COMMENT 'the branch to build of this pipelinerun',
    `git_ref`            varchar(128)                 DEFAULT NULL,
    `git_ref_type`       varchar(64)                  DEFAULT NULL,
    `git_commit`         varchar(128)                 DEFAULT NULL COMMENT 'the commit to build of this pipelinerun',
    `image_url`          varchar(256)                 DEFAULT NULL COMMENT 'image url',
    `last_config_commit` varchar(128)                 DEFAULT NULL COMMENT 'the last commit of cluster config',
    `config_commit`      varchar(128)                 DEFAULT NULL COMMENT 'the new commit of cluster config',
    `s3_bucket`          varchar(128)        NOT NULL DEFAULT '' COMMENT 's3 bucket to storage this pipelinerun log',
    `log_object`         varchar(258)        NOT NULL DEFAULT '' COMMENT 's3 object for log',
    `pr_object`          varchar(258)        NOT NULL DEFAULT '' COMMENT 's3 object for pipelinerun',
    `started_at`         datetime                     DEFAULT NULL COMMENT 'start time of this pipelinerun',
    `finished_at`        datetime                     DEFAULT NULL COMMENT 'finish time of this pipelinerun',
    `rollback_from`      bigint(20) unsigned          DEFAULT NULL COMMENT 'the pipelinerun id that this pipelinerun rollback from',
    `created_at`         datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`         datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `created_by`         bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'creator',
    PRIMARY KEY (`id`),
    KEY `idx_cluster_action` (`cluster_id`, `action`),
    KEY `idx_cluster_config_commit` (`cluster_id`, `config_commit`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- application region table
CREATE TABLE `tb_application_region`
(
    `id`               bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `application_id`   bigint(20) unsigned NOT NULL COMMENT 'application id',
    `environment_name` varchar(128)        NOT NULL DEFAULT '' COMMENT 'environment name',
    `region_name`      varchar(128)        NOT NULL DEFAULT '' COMMENT 'default deploy region of the environment',
    `created_at`       datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`       datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `created_by`       bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'creator',
    `updated_by`       bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'updater',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_application_environment` (`application_id`, `environment_name`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- tekton pipeline
CREATE TABLE `tb_pipeline`
(
    `id`             bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `pipelinerun_id` bigint(20) unsigned NOT NULL COMMENT 'pipelinerun id',
    `application`    varchar(64)         NOT NULL COMMENT 'application name',
    `cluster`        varchar(64)         NOT NULL COMMENT 'cluster name',
    `region`         varchar(16)         NOT NULL COMMENT 'region name',
    `pipeline`       varchar(16)         NOT NULL DEFAULT '' COMMENT 'pipeline name',
    `result`         varchar(16)         NOT NULL DEFAULT '' COMMENT 'result of the step, ok、failed or others',
    `duration`       int(16)             NOT NULL COMMENT 'duration',
    `started_at`     datetime            NOT NULL COMMENT 'start time of this pipelinerun',
    `finished_at`    datetime            NOT NULL COMMENT 'finish time of this pipelinerun',
    `created_at`     datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`     datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_region_application_created_at` (`region`, `application`, `created_at`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- tekton pipeline task
CREATE TABLE `tb_task`
(
    `id`             bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `pipelinerun_id` bigint(20) unsigned NOT NULL COMMENT 'pipelinerun id',
    `application`    varchar(64)         NOT NULL COMMENT 'application name',
    `cluster`        varchar(64)         NOT NULL COMMENT 'cluster name',
    `region`         varchar(16)         NOT NULL COMMENT 'region name',
    `pipeline`       varchar(16)         NOT NULL DEFAULT '' COMMENT 'pipeline name',
    `task`           varchar(16)         NOT NULL DEFAULT '' COMMENT 'task name',
    `result`         varchar(16)         NOT NULL DEFAULT '' COMMENT 'result of the step, ok or failed',
    `duration`       int(16)             NOT NULL COMMENT 'duration',
    `started_at`     datetime            NOT NULL COMMENT 'start time of this pipelinerun',
    `finished_at`    datetime            NOT NULL COMMENT 'finish time of this pipelinerun',
    `created_at`     datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`     datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_region_application_created_at` (`region`, `application`, `created_at`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- tekton task step
CREATE TABLE `tb_step`
(
    `id`             bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `pipelinerun_id` bigint(20) unsigned NOT NULL COMMENT 'pipelinerun id',
    `application`    varchar(64)         NOT NULL COMMENT 'application name',
    `cluster`        varchar(64)         NOT NULL COMMENT 'cluster name',
    `region`         varchar(16)         NOT NULL COMMENT 'region name',
    `pipeline`       varchar(16)         NOT NULL DEFAULT '' COMMENT 'pipeline name',
    `task`           varchar(16)         NOT NULL DEFAULT '' COMMENT 'task name',
    `step`           varchar(16)         NOT NULL DEFAULT '' COMMENT 'step name',
    `result`         varchar(16)         NOT NULL DEFAULT '' COMMENT 'result of the step, ok or failed',
    `duration`       int(16)             NOT NULL COMMENT 'duration',
    `started_at`     datetime            NOT NULL COMMENT 'start time of this pipelinerun',
    `finished_at`    datetime            NOT NULL COMMENT 'finish time of this pipelinerun',
    `created_at`     datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`     datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_region_application_created_at` (`region`, `application`, `created_at`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- oauth app table
CREATE TABLE `tb_oauth_app`
(
    `id`           bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `name`         varchar(128)                 DEFAULT NULL COMMENT 'short name of app client',
    `client_id`    varchar(128)                 DEFAULT NULL COMMENT 'oauth app client',
    `redirect_url` varchar(256)                 DEFAULT NULL COMMENT 'the authorization callback url',
    `home_url`     varchar(256)                 DEFAULT NULL COMMENT 'the oauth app home url',
    `description`  varchar(256)                 DEFAULT NULL COMMENT 'the desc of app',
    `app_type`     tinyint(1)          NOT NULL DEFAULT '1' COMMENT '1 for HorizonOAuthAPP, 2 for DirectOAuthAPP',
    `owner_type`   tinyint(1)          NOT NULL DEFAULT '1' COMMENT '1 for group, 2 for user',
    `owner_id`     bigint(20)                   DEFAULT NULL COMMENT 'group owner id',
    `created_at`   datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created_at',
    `created_by`   bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'creator',
    `updated_at`   datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `updated_by`   bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'updater',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_client_id` (`client_id`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- oauth client secret table
CREATE TABLE `tb_oauth_client_secret`
(
    `id`            bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `client_id`     varchar(256)                 DEFAULT NULL COMMENT 'oauth app client',
    `client_secret` varchar(256)                 DEFAULT NULL COMMENT 'oauth app secret',
    `created_at`    datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `created_by`    bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'creator',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_client_id_secret` (`client_id`, `client_secret`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- token table
CREATE TABLE `tb_token`
(
    `id`                     bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `client_id`              varchar(256)                 DEFAULT NULL COMMENT 'oauth app client',
    `redirect_uri`           varchar(256)                 DEFAULT NULL,
    `state`                  varchar(256)                 DEFAULT NULL COMMENT ' authorize_code state info',
    `code`                   varchar(256)        NOT NULL DEFAULT '' COMMENT 'private-token-code/authorize_code/access_token/refresh-token',
    `created_at`             datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `expires_in`             bigint(20)                   DEFAULT NULL,
    `scope`                  varchar(256)                 DEFAULT NULL,
    `user_or_robot_identity` varchar(256)                 DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_code` (`code`),
    KEY `idx_client_id` (`client_id`),
    KEY `idx_user_or_robot_identity` (`user_or_robot_identity`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- identity provider table
create table `tb_identity_provider`
(
    `id`                         bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `display_name`               varchar(128)        NOT NULL DEFAULT '' COMMENT 'name displayed on web',
    `name`                       varchar(128)        NOT NULL DEFAULT '' COMMENT 'name to generate index in db, unique',
    `avatar`                     varchar(256)        NOT NULL DEFAULT '' COMMENT 'link to avatar',
    `authorization_endpoint`     varchar(256)        NOT NULL DEFAULT '' COMMENT 'authorization endpoint of idp',
    `token_endpoint`             varchar(256)        NOT NULL DEFAULT '' COMMENT 'token endpoint of idp',
    `userinfo_endpoint`          varchar(256)        NOT NULL DEFAULT '' COMMENT 'userinfo endpoint of idp',
    `revocation_endpoint`        varchar(256)        NOT NULL DEFAULT '' COMMENT 'revocation endpoint of idp',
    `issuer`                     varchar(256)        NOT NULL DEFAULT '' COMMENT 'issuer of idp, generating discovery endpoint',
    `scopes`                     varchar(256)        NOT NULL DEFAULT '' COMMENT 'scopes when asking for authorization',
    `signing_algs`               varchar(256)        NOT NULL DEFAULT '' COMMENT 'algs for verifying signing',
    `token_endpoint_auth_method` varchar(256)        NOT NULL DEFAULT 'client_secret_sent_as_post' COMMENT 'how to carry client secret',
    `jwks`                       varchar(256)        NOT NULL DEFAULT '' COMMENT 'jwks endpoint, describe how to identify a token',
    `client_id`                  varchar(256)        NOT NULL DEFAULT '' COMMENT 'client id issued by idp',
    `client_secret`              varchar(256)        NOT NULL DEFAULT '' COMMENT 'client secret issued by idp',
    `created_at`                 datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'time of first creating',
    `updated_at`                 datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'time of last updating',
    `deleted_ts`                 bigint(20)                   DEFAULT '0' COMMENT 'deleted timestamp, 0 means not deleted',
    `created_by`                 bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'creator',
    `updated_by`                 bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'updater',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_name` (`name`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;
