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

alter table `horizon`.`application` rename to `horizon`.`tb_application`;
alter table `horizon`.`application_region` rename to `horizon`.`tb_application_region`;
alter table `horizon`.`cluster` rename to `horizon`.`tb_cluster`;
alter table `horizon`.`cluster_tag` rename to `horizon`.`tb_cluster_tag`;
alter table `horizon`.`cluster_template_schema_tag` rename to `horizon`.`tb_cluster_template_schema_tag`;
alter table `horizon`.`environment` rename to `horizon`.`tb_environment`;
alter table `horizon`.`environment_region` rename to `horizon`.`tb_environment_region`;
alter table `horizon`.`group` rename to `horizon`.`tb_group`;
alter table `horizon`.`harbor` rename to `horizon`.`tb_harbor`;
alter table `horizon`.`k8s_cluster` rename to `horizon`.`tb_k8s_cluster`;
alter table `horizon`.`member` rename to `horizon`.`tb_member`;
alter table `horizon`.`pipeline` rename to `horizon`.`tb_pipeline`;
alter table `horizon`.`pipelinerun` rename to `horizon`.`tb_pipelinerun`;
alter table `horizon`.`region` rename to `horizon`.`tb_region`;
alter table `horizon`.`step` rename to `horizon`.`tb_step`;
alter table `horizon`.`task` rename to `horizon`.`tb_task`;
alter table `horizon`.`template` rename to `horizon`.`tb_template`;
alter table `horizon`.`template_release` rename to `horizon`.`tb_template_release`;
alter table `horizon`.`user` rename to `horizon`.`tb_user`;
