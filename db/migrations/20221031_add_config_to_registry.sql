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

CREATE TABLE `tb_registry` (
  `id`                       bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `name`                     varchar(128) NOT NULL DEFAULT '' COMMENT 'name of the registry',
  `server`                   varchar(256) NOT NULL DEFAULT '' COMMENT 'registry server address',
  `token`                    varchar(512) NOT NULL DEFAULT '' COMMENT 'registry server token',
  `path`                     varchar(256) NOT NULL DEFAULT '' COMMENT 'path of image',
  `insecure_skip_tls_verify` tinyint(1) NOT NULL DEFAULT false COMMENT 'skip tls verify',
  `kind`                     varchar(256) NOT NULL DEFAULT 'harbor' COMMENT 'which kind registry it is',
  `created_at`               datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`               datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_ts`               bigint(20) DEFAULT '0' COMMENT 'deleted timestamp, 0 means not deleted',
  `created_by`               bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'creator',
  `updated_by`               bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'updater',
  PRIMARY KEY (`id`)
) ENGINE = InnoDB 
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

INSERT INTO tb_registry (`id`, `name`, `server`, `token`, `insecure_skip_tls_verify`)
(SELECT `id`, `name`, `server`, `token`, true from tb_harbor);

ALTER TABLE tb_region add column `registry_id` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'registry id';
UPDATE tb_region set `registry_id` = `harbor_id`;