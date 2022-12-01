-- tb_event is used to record key events
CREATE TABLE `tb_event` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `req_id` varchar(256) NOT NULL DEFAULT '' COMMENT 'request id that caused this event to be generated',
  `resource_type` varchar(256) NOT NULL DEFAULT '' COMMENT 'type of resource, currently support: applications, clusters',
  `resource_id` varchar(256) NOT NULL DEFAULT '' COMMENT 'resource id that caused this event to be generated',
  `event_type` varchar(256) NOT NULL DEFAULT '' COMMENT 'type of event',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'creation timestamp',
  `created_by` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'creator',
  PRIMARY KEY (`id`),
  KEY `idx_req_id` (`req_id`),
  KEY `idx_resource_action` (`resource_id`, `resource_type`, `event_type`)
) ENGINE = InnoDB AUTO_INCREMENT = 1 DEFAULT CHARSET = utf8mb4;
-- tb_event_cursor is used to record where the event is consumed
CREATE TABLE `tb_event_cursor` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `position` bigint(20) NOT NULL DEFAULT 0 COMMENT 'id of the last processed event',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'creation timestamp',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'update time',
  PRIMARY KEY (`id`)
) ENGINE = InnoDB AUTO_INCREMENT = 1 DEFAULT CHARSET = utf8mb4;
CREATE TABLE `tb_webhook` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `enabled` tinyint(1) NOT NULL DEFAULT true COMMENT 'whether this webhook is enabled',
  `url` text NOT NULL COMMENT 'url of target server',
  `ssl_verify_enabled` tinyint(1) NOT NULL DEFAULT false COMMENT 'whether to skip ssl verification',
  `description` varchar(256) NOT NULL DEFAULT '' COMMENT 'description',
  `secret` text NOT NULL COMMENT 'used as authentication credentials',
  `triggers` text NOT NULL COMMENT 'event type list to trigger this webhook',
  `resource_type` varchar(256) NOT NULL DEFAULT '' COMMENT 'type of resource, currently support: groups, application, clusters',
  `resource_id` bigint(20) NOT NULL DEFAULT 0 COMMENT 'id of resource',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'creation timestamp',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'update time',
  `created_by` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'creator',
  `updated_by` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'updater',
  PRIMARY KEY (`id`),
  KEY `idx_resource` (`resource_id`, `resource_type`)
) ENGINE = InnoDB AUTO_INCREMENT = 1 DEFAULT CHARSET = utf8mb4;
CREATE TABLE `tb_webhook_log` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `webhook_id` bigint(20) unsigned NOT NULL COMMENT 'id of webhook',
  `event_id` bigint(20) unsigned NOT NULL COMMENT 'id of event',
  `url` text NOT NULL COMMENT 'url of target server',
  `request_headers` text NOT NULL COMMENT 'request headers',
  `request_data` text NOT NULL COMMENT 'request data to be marshaled as body',
  `response_headers` text NOT NULL COMMENT 'response headers',
  `response_body` text NOT NULL COMMENT 'response body of',
  `status` varchar(256) NOT NULL COMMENT 'status, currently support: waiting, failed, success',
  `error_message` text NOT NULL COMMENT 'error message',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'creation timestamp',
  `created_by` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'creator',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'updater',
  PRIMARY KEY (`id`),
  KEY `idx_webhook_id_status` (`webhook_id`, `status`),
  KEY `idx_event_id` (`event_id`)
) ENGINE = InnoDB AUTO_INCREMENT = 1 DEFAULT CHARSET = utf8mb4;