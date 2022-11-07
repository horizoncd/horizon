ALTER TABLE
    tb_user
ADD
    column `created_by` bigint(20) unsigned NOT NULL DEFAULT 0,
ADD
    column `user_type` tinyint(1) unsigned NOT NULL DEFAULT 0;

ALTER TABLE
    tb_token
ADD
    column `name` varchar(64) NOT NULL DEFAULT '',
ADD
    column `created_by` bigint(20) unsigned NOT NULL DEFAULT '0';