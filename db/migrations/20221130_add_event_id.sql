ALTER TABLE tb_pipelinerun
    ADD column `ci_event_id` varchar(36) NOT NULL DEFAULT '',
    ADD INDEX idx_ci_event_id (`ci_event_id`);
