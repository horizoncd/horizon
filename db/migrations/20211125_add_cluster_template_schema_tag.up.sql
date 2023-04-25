-- Copyright © 2023 Horizoncd.
--
-- Licensed under the Apache License, Version 2.0 (the "License");
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
--     http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an "AS IS" BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.

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
    KEY          `idx_key` (`tag_key`),
    UNIQUE KEY `idx_cluster_id_key` (`cluster_id`, `tag_key`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;
