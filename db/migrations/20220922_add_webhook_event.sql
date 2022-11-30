CREATE TABLE `tb_event` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `req_id` varchar(256) NOT NULL DEFAULT '',
  `resource_type` varchar(256) NOT NULL DEFAULT '' COMMENT 'currently support: groups, applications, clusters',
  `resource_id` varchar(256) NOT NULL DEFAULT '',
  `action` varchar(256) NOT NULL DEFAULT '',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `created_by` bigint(20) unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_req_id` (`req_id`),
  KEY `idx_resource_action` (`resource_id`, `resource_type`, `action`)
) ENGINE = InnoDB AUTO_INCREMENT = 1 DEFAULT CHARSET = utf8mb4;

CREATE TABLE `tb_event_cursor` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `position` bigint(20) NOT NULL DEFAULT 0,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE = InnoDB AUTO_INCREMENT = 1 DEFAULT CHARSET = utf8mb4;

CREATE TABLE `tb_webhook` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `enabled` tinyint(1) NOT NULL DEFAULT true,
  `url` text NOT NULL,
  `ssl_verify_enabled` tinyint(1) NOT NULL DEFAULT false,
  `description` varchar(256) NOT NULL DEFAULT '',
  `secret` text NOT NULL,
  `triggers` text NOT NULL,
  `resource_type` varchar(256) NOT NULL DEFAULT '',
  `resource_id` bigint(20) NOT NULL DEFAULT 0,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `created_by` bigint(20) unsigned NOT NULL DEFAULT 0,
  `updated_by` bigint(20) unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`)
) ENGINE = InnoDB AUTO_INCREMENT = 1 DEFAULT CHARSET = utf8mb4;

CREATE TABLE `tb_webhook_log` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `webhook_id` bigint(20) unsigned NOT NULL,
  `event_id` bigint(20) unsigned NOT NULL,
  `url` text NOT NULL,
  `request_headers` text NOT NULL,
  `request_data` text NOT NULL,
  `response_headers` text NOT NULL,
  `response_body` text NOT NULL,
  `status` varchar(256) NOT NULL,
  `error_message` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `created_by` bigint(20) unsigned NOT NULL DEFAULT 0,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_webhook_id_status` (`webhook_id`, `status`),
  KEY `idx_event_id` (`event_id`)
) ENGINE = InnoDB AUTO_INCREMENT = 1 DEFAULT CHARSET = utf8mb4;