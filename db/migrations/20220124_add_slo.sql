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

-- tekton pipeline
CREATE TABLE `pipeline`
(
    `id`             bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `pipelinerun_id` bigint(20) unsigned NOT NULL COMMENT 'pipelinerun id',
    `application`    varchar(64)         NOT NULL COMMENT 'application name',
    `cluster`        varchar(64)         NOT NULL COMMENT 'cluster name',
    `region`         varchar(16)         NOT NULL COMMENT 'region name',
    `pipeline`       varchar(16)         NOT NULL DEFAULT '' COMMENT 'pipeline name',
    `result`         varchar(16)         NOT NULL DEFAULT '' COMMENT 'result of the step, ok、failed or others',
    `duration`       int(16)             NOT NULL COMMENT 'duration',
    `started_at`     datetime            NOT NULL COMMENT 'start time of this pipelinerun',
    `finished_at`    datetime            NOT NULL COMMENT 'finish time of this pipelinerun',
    `created_at`     datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`     datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_region_application_created_at` (`region`, `application`, `created_at`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- tekton pipeline task
CREATE TABLE `task`
(
    `id`             bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `pipelinerun_id` bigint(20) unsigned NOT NULL COMMENT 'pipelinerun id',
    `application`    varchar(64)         NOT NULL COMMENT 'application name',
    `cluster`        varchar(64)         NOT NULL COMMENT 'cluster name',
    `region`         varchar(16)         NOT NULL COMMENT 'region name',
    `pipeline`       varchar(16)         NOT NULL DEFAULT '' COMMENT 'pipeline name',
    `task`           varchar(16)         NOT NULL DEFAULT '' COMMENT 'task name',
    `result`         varchar(16)         NOT NULL DEFAULT '' COMMENT 'result of the step, ok or failed',
    `duration`       int(16)             NOT NULL COMMENT 'duration',
    `started_at`     datetime            NOT NULL COMMENT 'start time of this pipelinerun',
    `finished_at`    datetime            NOT NULL COMMENT 'finish time of this pipelinerun',
    `created_at`     datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`     datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_region_application_created_at` (`region`, `application`, `created_at`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

-- tekton task step
CREATE TABLE `step`
(
    `id`             bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `pipelinerun_id` bigint(20) unsigned NOT NULL COMMENT 'pipelinerun id',
    `application`    varchar(64)         NOT NULL COMMENT 'application name',
    `cluster`        varchar(64)         NOT NULL COMMENT 'cluster name',
    `region`         varchar(16)         NOT NULL COMMENT 'region name',
    `pipeline`       varchar(16)         NOT NULL DEFAULT '' COMMENT 'pipeline name',
    `task`           varchar(16)         NOT NULL DEFAULT '' COMMENT 'task name',
    `step`           varchar(16)         NOT NULL DEFAULT '' COMMENT 'step name',
    `result`         varchar(16)         NOT NULL DEFAULT '' COMMENT 'result of the step, ok or failed',
    `duration`       int(16)             NOT NULL COMMENT 'duration',
    `started_at`     datetime            NOT NULL COMMENT 'start time of this pipelinerun',
    `finished_at`    datetime            NOT NULL COMMENT 'finish time of this pipelinerun',
    `created_at`     datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`     datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_region_application_created_at` (`region`, `application`, `created_at`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;
