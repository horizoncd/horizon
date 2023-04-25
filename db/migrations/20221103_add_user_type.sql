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

ALTER TABLE
    tb_user
ADD
    column `created_by` bigint(20) unsigned NOT NULL DEFAULT 0,
ADD
    column `user_type` tinyint(1) unsigned NOT NULL DEFAULT 0 COMMENT 'the option type is: 0 (common user), 1(robot user)';

ALTER TABLE
    tb_token
ADD
    column `name` varchar(64) NOT NULL DEFAULT '',
ADD
    column `user_id` bigint(20) unsigned NOT NULL DEFAULT '0',
ADD
    column `created_by` bigint(20) unsigned NOT NULL DEFAULT '0';