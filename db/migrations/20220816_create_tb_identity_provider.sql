create table `tb_identity_provider`
(
    `id`                            bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `display_name`                  varchar(128) NOT NULL DEFAULT '' COMMENT 'name displayed on web',
    `name`                          varchar(128) NOT NULL DEFAULT '' COMMENT 'name to generate index in db, unique',
    `avatar`                        varchar(256) NOT NULL DEFAULT '' COMMENT 'link to avatar',
    `authorization_endpoint`        varchar(256) NOT NULL DEFAULT '' COMMENT 'authorization endpoint of idp',
    `token_endpoint`                varchar(256) NOT NULL DEFAULT '' COMMENT 'token endpoint of idp',
    `userinfo_endpoint`             varchar(256) NOT NULL DEFAULT '' COMMENT 'userinfo endpoint of idp',
    `revocation_endpoint`           varchar(256) NOT NULL DEFAULT '' COMMENT 'revocation endpoint of idp',
    `issuer`                        varchar(256) NOT NULL DEFAULT '' COMMENT 'issuer of idp, generating discovery endpoint',
    `scopes`                        varchar(256) NOT NULL DEFAULT '' COMMENT 'scopes when asking for authorization',
    `token_endpoint_auth_method`    varchar(256) NOT NULL DEFAULT 'client_secret_sent_as_post' COMMENT 'how to carry client secret',
    `jwks`                          varchar(256) NOT NULL DEFAULT '' COMMENT 'jwks endpoint, describe how to identify a token',
    `client_id`                     varchar(256) NOT NULL DEFAULT '' COMMENT 'client id issued by idp',
    `client_secret`                 varchar(256) NOT NULL DEFAULT '' COMMENT 'client secret issued by idp',
    `created_at`                    datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'time of first creating',
    `updated_at`                    datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'time of last updating',
    `deleted_ts`                    bigint(20)   DEFAULT '0' COMMENT 'deleted timestamp, 0 means not deleted',
    `created_by`                    bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'creator',
    `updated_by`                    bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'updater',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_name` (`name`)
)   ENGINE = InnoDB
    AUTO_INCREMENT = 1
    DEFAULT CHARSET = utf8mb4;