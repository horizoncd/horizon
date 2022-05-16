-- Token Table
CREATE TABLE `tb_token`
(
    `id`        bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `client_id`  varchar(256)  COMMENT 'oauth app client',
    `redirect_uri` varchar(256),
    `state`    varchar(256) COMMENT ' authorize_code state info',

    `code` varchar(256) NOT NULL COMMENT 'private-token-code/authorize_code/access_token/refresh-token',
    `created_at` datetime  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `expire_in`  bigint(20),
    `scope` varchar(256),
    `user_or_robot_identity` varchar(256),
    `created_by`     bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'creator',
    `updated_by`     bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'updater',
    PRIMARY KEY (`id`)
    UNIQUE KEY `idx_code` (`code`)
    KEY `idx_client_id` (`client_id`)
    KEY `idx_user_or_robot_identity` (`user_or_robot_identity`)
)

-- Oauth Client Table
CREATE table `tb_oauth_client` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `client_id`  varchar(256)  COMMENT 'oauth app client',
    `redirect_uri` varchar(256) COMMNET 'the authorization callback url',
    `home_url`   varchar(256) COMMNET 'the oauth app home url',
    `created_at` datetime  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `created_by`     bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'creator',
    `updated_by`     bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'updater',
     PRIMARY KEY (`id`)
     UNIQUE KEY `idx_client_id` (`client_id`)
)

-- Oauth Client Secret Table
CREATE table `tb_oauth_client_secret` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `client_id`  varchar(256)  COMMENT 'oauth app client',
    `client_secret`  varchar(256)  COMMENT 'oauth app secret',
    `created_at` datetime  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `created_by`     bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'creator',
    PRIMARY KEY (`id`)
    UNIQUE KEY `idx_client_id_secret` (`client_id`, `client_secret`)
)