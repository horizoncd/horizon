-- group table
CREATE TABLE `group` (
     `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
     `name` varchar(128) NOT NULL DEFAULT '',
     `full_name` varchar(512) NOT NULL DEFAULT 'name from roots, for example: 1 / 2 / 3',
     `path` varchar(256) NOT NULL DEFAULT '' COMMENT 'path from roots, for example: a/b/c',
     `description` varchar(256) DEFAULT NULL,
     `visibility_level` varchar(16) NOT NULL COMMENT 'public or private',
     `parent_id` int(11) NOT NULL DEFAULT -1 COMMENT 'ID of the parent group',
     `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
     `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
     `deleted_at` datetime DEFAULT NULL,
     PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

-- user table
CREATE TABLE `user` (
    `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `name` varchar(64) NOT NULL DEFAULT '' COMMENT '名称',
    `email` varchar(64) NOT NULL DEFAULT '' COMMENT '邮件地址',
    `phone` varchar(32) DEFAULT NULL COMMENT '电话号码',
    `oidc_id` varchar(64) NOT NULL COMMENT 'oidc系统ID，比如oidc_type为netease的话，此字段为工号',
    `oidc_type` varchar(64) NOT NULL COMMENT 'oidc类型，当前主要为netease',
    `admin`  tinyint(1) NOT NULL COMMENT '是否是管理员，0-否，1-是',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` datetime DEFAULT NULL COMMENT '删除时间，未删除时为null',
    PRIMARY KEY (`id`),
    KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;