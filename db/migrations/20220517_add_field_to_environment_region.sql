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

-- set region to be default
ALTER TABLE horizon.tb_environment_region ADD is_default tinyint(1) NOT NULL DEFAULT 0 COMMENT '0 means not default region, 1 means default region' AFTER region_name;

-- this field is not used anymore
Alter TABLE horizon.tb_region DROP k8s_cluster_id;

-- all fields has been moved to tb_region, this table can be dropped
DROP TABLE horizon.tb_k8s_cluster;

-- add name field to harbor
ALTER TABLE horizon.tb_harbor ADD name varchar(128) NOT NULL DEFAULT '' COMMENT 'name of the harbor registry' AFTER id;
