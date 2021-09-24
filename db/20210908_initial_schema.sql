-- group table
CREATE TABLE `group` (
    `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
    `name` varchar(128) NOT NULL DEFAULT '',
    `path` varchar(32) NOT NULL DEFAULT '',
    `description` varchar(256) DEFAULT NULL,
    `visibility_level` varchar(16) NOT NULL COMMENT 'public or private',
    `parent_id` int(11) NOT NULL DEFAULT -1 COMMENT 'ID of the parent group',
    `traversal_ids` varchar(32) NOT NULL DEFAULT '' COMMENT 'ID path from the root, like 1,2,3',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` datetime DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_parent_id_name` (`parent_id`, `name`),
    UNIQUE KEY `idx_parent_id_path` (`parent_id`, `path`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

-- user table
CREATE TABLE `user` (
    `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
    `name` varchar(64) NOT NULL DEFAULT '',
    `email` varchar(64) NOT NULL DEFAULT '',
    `phone` varchar(32) DEFAULT NULL COMMENT,
    `oidc_id` varchar(64) NOT NULL COMMENT 'oidc id, which is a unique key in oidc system.',
    `oidc_type` varchar(64) NOT NULL COMMENT 'oidc type, such as google, github, gitlab etc.',
    `admin`  tinyint(1) NOT NULL COMMENT 'is system admin，0-false，1-true',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` datetime DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_deleted_at` (`deleted_at`),
    UNIQUE KEY `idx_name`(`name`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

-- gitlab table
CREATE TABLE `gitlab` (
    `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
    `name` varchar(64) NOT NULL DEFAULT '' COMMENT 'the name of gitlab',
    `url` varchar(128) NOT NULL DEFAULT '' COMMENT 'gitlab base url',
    `token` varchar(128) NOT NULL DEFAULT '' COMMENT 'gitlab token',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` datetime DEFAULT NULL,
    `created_by` varchar(64) NOT NULL DEFAULT '' COMMENT 'creator',
    `updated_by` varchar(64) NOT NULL DEFAULT '' COMMENT 'updater',
    PRIMARY KEY (`id`),
    KEY `idx_deleted_at` (`deleted_at`),
    UNIQUE KEY `idx_name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

-- template table
CREATE TABLE `template` (
    `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
    `name` varchar(64) NOT NULL DEFAULT '' COMMENT 'the name of template',
    `description` varchar(256) NULL COMMENT 'the template description',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` datetime DEFAULT NULL,
    `created_by` varchar(64) NOT NULL DEFAULT '' COMMENT 'creator',
    `updated_by` varchar(64) NOT NULL DEFAULT '' COMMENT 'updater',
    PRIMARY KEY (`id`),
    KEY `idx_deleted_at` (`deleted_at`),
    UNIQUE KEY `idx_name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

-- template release table
CREATE TABLE `template_release` (
    `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
    `template_name` varchar(64) NOT NULL COMMENT 'the name of template',
    `name` varchar(64) NOT NULL DEFAULT '' COMMENT 'the name of template release',
    `description` varchar(256) NOT NULL COMMENT 'description about this template release',
    `gitlab_name` varchar(64) NOT NULL COMMENT 'the name of gitlab',
    `gitlab_project` varchar(256) NOT NULL COMMENT 'project ID or relative path in gitlab',
    `recommended` tinyint(1) NOT NULL COMMENT 'is the most recommended template, 0-false, 1-true',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` datetime DEFAULT NULL,
    `created_by` varchar(64) NOT NULL DEFAULT '' COMMENT 'creator',
    `updated_by` varchar(64) NOT NULL DEFAULT '' COMMENT 'updater',
    PRIMARY KEY (`id`),
    KEY `idx_deleted_at` (`deleted_at`),
    UNIQUE KEY `idx_template_name_name` (`template_name`, `name`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

-- application table
CREATE TABLE `application` (
    `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
    `group_id` int(11) unsigned NOT NULL COMMENT 'group id',
    `name` varchar(64) NOT NULL DEFAULT '' COMMENT 'the name of application',
    `description` varchar(256) DEFAULT NULL COMMENT 'the description of application',
    `priority` varchar(16) NOT NULL DEFAULT 'P3' COMMENT 'the priority of application',
    `git_url` varchar(128) DEFAULT NULL COMMENT 'git repo url',
    `git_subfolder` varchar(128) DEFAULT NULL COMMENT 'git repo subfolder',
    `git_branch` varchar(128) DEFAULT NULL COMMENT 'git default branch',
    `template` varchar(64) NOT NULL COMMENT 'template name',
    `template_release` varchar(64) NOT NULL COMMENT 'template release',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` datetime DEFAULT NULL,
    `created_by` varchar(64) NOT NULL DEFAULT '' COMMENT 'creator',
    `updated_by` varchar(64) NOT NULL DEFAULT '' COMMENT 'updater',
    PRIMARY KEY (`id`),
    KEY `idx_deleted_at` (`deleted_at`),
    UNIQUE KEY `idx_name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;