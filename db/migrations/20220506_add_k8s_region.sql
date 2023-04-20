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

-- merge k8s_cluster to region
ALTER TABLE tb_region ADD server varchar(256) NULL COMMENT 'k8s server url' AFTER display_name;
ALTER TABLE tb_region ADD certificate text NULL COMMENT 'k8s kube config' AFTER server;
ALTER TABLE tb_region ADD ingress_domain text NULL COMMENT 'k8s ingress domain' AFTER certificate;

-- add default_region to environment
ALTER TABLE tb_environment ADD default_region varchar(128) NULL COMMENT 'default region of the environment' AFTER display_name;
