ALTER TABLE horizon.tb_pipelinerun
    ADD column `ci_event_id` varchar(36) NOT NULL DEFAULT '';

CREATE UNIQUE INDEX idx_event_id ON `horizon`.`tb_pipelinerun` (`ci_event_id`);
