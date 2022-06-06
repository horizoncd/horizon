-- Token Table
CREATE TABLE `tb_token`
(
    `id`        bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `client_id`  varchar(256)  COMMENT 'oauth app client',
    `redirect_uri` varchar(256),
    `state`    varchar(256) COMMENT ' authorize_code state info',
    `code` varchar(256) NOT NULL COMMENT 'private-token-code/authorize_code/access_token/refresh-token',
    `created_at` datetime  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `expires_in`  bigint(20),
    `scope` varchar(256),
    `user_or_robot_identity` varchar(256),
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_code` (`code`),
    KEY `idx_client_id` (`client_id`),
    KEY `idx_user_or_robot_identity` (`user_or_robot_identity`)
)

-- Oauth app Table
CREATE table `tb_oauth_app` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `name`       varchar(128) COMMENT 'short name of app client',
    `client_id`  varchar(128)  COMMENT 'oauth app client',
    `redirect_url` varchar(256) COMMENT 'the authorization callback url',
    `home_url`   varchar(256) COMMENT 'the oauth app home url',
    `desc`      varchar(256),
    `app_type`     tinyint(1)  NOT NULL COMMENT '1 for HorizonOAuthAPP, 2 for DirectOAuthAPP',
    `owner_type` tinyint(1) NOT NULL COMMENT '1 for group, 2 for user',
    `owner_id`   bigint(20) NOT NULL,
    `created_at` datetime  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `created_by`     bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'creator',
    `updated_by`     bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'updater',
     PRIMARY KEY (`id`),
     UNIQUE KEY `idx_client_id` (`client_id`)
)

-- Oauth Client Secret Table
CREATE table `tb_oauth_client_secret` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `client_id`  varchar(256)  COMMENT 'oauth app client',
    `client_secret`  varchar(256)  COMMENT 'oauth app secret',
    `created_at` datetime  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `created_by`     bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'creator',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_client_id_secret` (`client_id`, `client_secret`)
)

-- Horizon app table
CREATE table `tb_horizon_app` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `name` varchar(128) NOT NULL COMMENT 'name of horizon app',
    `desc` varchar(128) 'desc of horizon app',
    `home_url`   varchar(256) COMMNET 'the oauth app home url',
    `oauth_app_id` bigint(20) unsigned NOT NULL COMMENT 'the oauth app of horizon app',

    `owner_type` tinyint(1) NOT NULL COMMENT '1 for group, 2 for user',
    `owner_id`   bigint(20) NOT NULL,

    `created_at` datetime  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `created_by`     bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'creator',
    `updated_by`     bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'updater',
)

-- Horizon permission table
CREATE table `tb_horizon_app_permission` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `horizon_app_id` bigint(20) NOT NULL,
    `permissions` JSON  COMMENT 'Horizon app permission list',
    `created_at` datetime  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `created_by`     bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'creator',
    `updated_by`     bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'updater',
)