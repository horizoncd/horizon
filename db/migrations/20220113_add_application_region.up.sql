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

-- application region table
CREATE TABLE `application_region`
(
    `id`               bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `application_id`   bigint(20) unsigned NOT NULL COMMENT 'application id',
    `environment_name` varchar(128) NOT NULL DEFAULT '' COMMENT 'environment name',
    `region_name`      varchar(128) NOT NULL DEFAULT '' COMMENT 'default deploy region of the environment',
    `created_at`       datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`       datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `created_by`       bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'creator',
    `updated_by`       bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'updater',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_application_environment` (`application_id`, `environment_name`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;
