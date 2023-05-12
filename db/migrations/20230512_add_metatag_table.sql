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

-- metatag table
CREATE TABLE `tb_metatag`
(
    `tag_key`            varchar(64)  NOT NULL DEFAULT '' comment 'key of the metatag',
    `tag_value`          varchar(128) NOT NULL DEFAULT '' comment 'value of the metatag',
    `tag_value_identity` varchar(64)  NOT NULL DEFAULT '' comment 'identity of value of the metatag',
    `created_at`         datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`         datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY `idx_key_value` (`tag_key`, `tag_value`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;
