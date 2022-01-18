# 流水线Pipeline结果
CREATE TABLE `PipelineResult`
(
    `id`             bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `pipelinerun_id` bigint(20) unsigned NOT NULL COMMENT 'pipelinerun id',
    `application`    varchar(64)         NOT NULL COMMENT 'application name',
    `cluster`        varchar(64)         NOT NULL COMMENT 'cluster name',
    `environment`    varchar(16)         NOT NULL COMMENT 'environment name',
    `pipeline`       varchar(16)         NOT NULL DEFAULT '' COMMENT 'pipeline name',
    `result`         varchar(16)         NOT NULL DEFAULT '' COMMENT 'result of the step, ok or failed',
    `started_at`     datetime            NOT NULL COMMENT 'start time of this pipelinerun',
    `finished_at`    datetime            NOT NULL COMMENT 'finish time of this pipelinerun',
    `created_at`     datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`     datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_created_at` (`created_at`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

# 流水线中各个task的结果
CREATE TABLE `TaskResult`
(
    `id`                 bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `pipeline_result_id` bigint(20) unsigned NOT NULL COMMENT 'pipeline result id',
    `application`        varchar(64)         NOT NULL COMMENT 'application name',
    `cluster`            varchar(64)         NOT NULL COMMENT 'cluster name',
    `environment`        varchar(16)         NOT NULL COMMENT 'environment name',
    `pipeline`           varchar(16)         NOT NULL DEFAULT '' COMMENT 'pipeline name',
    `task`               varchar(16)         NOT NULL DEFAULT '' COMMENT 'task name',
    `result`             varchar(16)         NOT NULL DEFAULT '' COMMENT 'result of the step, ok or failed',
    `started_at`         datetime            NOT NULL COMMENT 'start time of this pipelinerun',
    `finished_at`        datetime            NOT NULL COMMENT 'finish time of this pipelinerun',
    `created_at`         datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`         datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_created_at` (`created_at`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;

# 流水线task中各个step的结果
CREATE TABLE `StepResult`
(
    `id`             bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `task_result_id` bigint(20) unsigned NOT NULL COMMENT 'task result id',
    `application`    varchar(64)         NOT NULL COMMENT 'application name',
    `cluster`        varchar(64)         NOT NULL COMMENT 'cluster name',
    `environment`    varchar(16)         NOT NULL COMMENT 'environment name',
    `pipeline`       varchar(16)         NOT NULL DEFAULT '' COMMENT 'pipeline name',
    `task`           varchar(16)         NOT NULL DEFAULT '' COMMENT 'task name',
    `step`           varchar(16)         NOT NULL DEFAULT '' COMMENT 'step name',
    `result`         varchar(16)         NOT NULL DEFAULT '' COMMENT 'result of the step, ok or failed',
    `started_at`     datetime            NOT NULL COMMENT 'start time of this pipelinerun',
    `finished_at`    datetime            NOT NULL COMMENT 'finish time of this pipelinerun',
    `created_at`     datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`     datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_created_at` (`created_at`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4;
