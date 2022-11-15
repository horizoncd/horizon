ALTER TABLE
    tb_user
ADD
    column `created_by` bigint(20) unsigned NOT NULL DEFAULT 0,
ADD
    column `user_type` tinyint(1) unsigned NOT NULL DEFAULT 0 COMMENT 'the option type is: 0 (common user), 1(robot user)';

ALTER TABLE
    tb_token
ADD
    column `name` varchar(64) NOT NULL DEFAULT '',
ADD
    column `user_id` bigint(20) unsigned NOT NULL DEFAULT '0',
ADD
    column `created_by` bigint(20) unsigned NOT NULL DEFAULT '0';