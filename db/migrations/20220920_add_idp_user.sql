create table `tb_idp_user`
(
    `id`                            bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `sub`                           varchar(256) NOT NULL DEFAULT '' COMMENT 'user id in idp',
    `idp_id`                        bigint(20) NOT NULL DEFAULT 0 COMMENT 'refer to tb_identify_provider',
    `user_id`                       bigint(20) NOT NULL DEFAULT 0 COMMENT 'refer to tb_user',
    `name`                          varchar(256) NOT NULL DEFAULT '' COMMENT 'user name from idp',
    `email`                         varchar(256) NOT NULL DEFAULT '' COMMENT 'user email from idp',
    `deletable`                     bool NOT NULL DEFAULT false COMMENT 'whether this link can be deleted',
    `created_at`                    datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'time of first creating',
    `updated_at`                    datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'time of last updating',
    `deleted_ts`                    bigint(20)   DEFAULT '0' COMMENT 'deleted timestamp, 0 means not deleted',
    `created_by`                    bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'creator',
    `updated_by`                    bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'updater',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uni_idx_idp_sub` (`idp_id`, `sub`, `deleted_ts`)
)   ENGINE = InnoDB
    AUTO_INCREMENT = 1
    DEFAULT CHARSET = utf8mb4;

alter table tb_user drop index idx_name;
alter table tb_user drop index idx_email;

alter table tb_user add column `banned` bool NOT NULL DEFAULT false COMMENT 'whether user is banned';
insert into tb_idp_user( `sub`,`idp_id`,`user_id`,`name`,`email`,`deletable`,`created_at`,`updated_at`,`deleted_ts`,`created_by`,`updated_by`) select u.`name`,`idp`.`id`, `u`.`id`, `u`.`name`, `u`.`email`, false, now(), now(), 0, 0, 0 from tb_user as u join (select * from tb_identity_provider where `name` = "netease") as idp