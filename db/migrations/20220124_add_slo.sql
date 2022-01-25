-- tekton pipeline
CREATE TABLE `Pipeline`
(
    `id`             bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `pipelinerun_id` bigint(20) unsigned NOT NULL COMMENT 'pipelinerun id',
    `application`    varchar(64)         NOT NULL COMMENT 'application name',
    `cluster`        varchar(64)         NOT NULL COMMENT 'cluster name',
    `region`         bigint(20)          NOT NULL COMMENT 'region name',
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
CREATE TABLE `Task`
(
    `id`             bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `pipelinerun_id` bigint(20) unsigned NOT NULL COMMENT 'pipelinerun id',
    `application`    varchar(64)         NOT NULL COMMENT 'application name',
    `cluster`        varchar(64)         NOT NULL COMMENT 'cluster name',
    `region`         bigint(20)          NOT NULL COMMENT 'region name',
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
CREATE TABLE `Step`
(
    `id`             bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `pipelinerun_id` bigint(20) unsigned NOT NULL COMMENT 'pipelinerun id',
    `application`    varchar(64)         NOT NULL COMMENT 'application name',
    `cluster`        varchar(64)         NOT NULL COMMENT 'cluster name',
    `region`         bigint(20)          NOT NULL COMMENT 'region name',
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
