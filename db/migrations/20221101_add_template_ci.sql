ALTER TABLE tb_template
    ADD without_ci tinyint(1) NOT NULL DEFAULT 0 COMMENT 'without_ci configuration, 0 means with ci';