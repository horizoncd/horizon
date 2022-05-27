alter table horizon.tb_group
    add region_selector varchar(512) default '' not null COMMENT 'used for filtering regions' after traversal_ids;

alter table horizon.tb_region
    add `disabled` TINYINT(1) default 0 not null COMMENT '0 means not disabled, 1 means disabled' after harbor_id;

-- this field is not used anymore
Alter TABLE horizon.tb_cluster DROP environment_region_id;

