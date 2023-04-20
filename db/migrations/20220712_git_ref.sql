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

ALTER TABLE horizon.tb_pipelinerun ADD git_ref varchar(128) NULL after git_branch;
ALTER TABLE horizon.tb_cluster ADD git_ref varchar(128) NULL after git_branch;
ALTER TABLE horizon.tb_application ADD git_ref varchar(128) NULL after git_branch;

ALTER TABLE horizon.tb_pipelinerun ADD git_ref_type varchar(64) NULL after git_ref;
ALTER TABLE horizon.tb_cluster ADD git_ref_type varchar(64) NULL after git_ref;
ALTER TABLE horizon.tb_application ADD git_ref_type varchar(64) NULL after git_ref;

UPDATE horizon.tb_cluster set git_ref_type = 'branch';
UPDATE horizon.tb_application set git_ref_type = 'branch';
UPDATE horizon.tb_pipelinerun set git_ref_type = 'branch';

UPDATE horizon.tb_cluster set git_ref = git_branch;
UPDATE horizon.tb_application set git_ref = git_branch;
UPDATE horizon.tb_pipelinerun set git_ref = git_branch;