ALTER TABLE tb_environment
    ADD auto_free tinyint(1) NOT NULL DEFAULT 0 COMMENT 'auto free configuration, 0 means disabled' AFTER default_region;
ALTER TABLE tb_cluster
    ADD expire_seconds bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'expiration seconds, 0 means permanent' AFTER status;