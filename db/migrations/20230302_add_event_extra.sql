ALTER TABLE tb_event
    ADD COLUMN extra varchar(255) NOT NULL DEFAULT ''
    COMMENT 'extra infos to describe the event'