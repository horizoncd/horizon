alter table horizon.tb_group
    add region_selector varchar(512) default '' not null COMMENT 'used for filtering kubernetes' after traversal_ids;

alter table horizon.tb_region
    add `disabled` TINYINT(1) default 0 not null COMMENT '0 means not disabled, 1 means disabled' after harbor_id;

-- this field is not used anymore
Alter TABLE horizon.tb_cluster
    DROP environment_region_id;

-- initialize data
update horizon.tb_group
set region_selector = '- key: cloudnative.music.netease.com/group
  values:
  - default
  operator: in
';

insert into horizon.tb_tag(resource_id, resource_type, tag_key, tag_value)
values
(9, 'kubernetes', 'cloudnative.music.netease.com/group', 'default'),
(5, 'kubernetes', 'cloudnative.music.netease.com/group', 'default'),
(2, 'kubernetes', 'cloudnative.music.netease.com/group', 'default'),
(7, 'kubernetes', 'cloudnative.music.netease.com/group', 'default'),
(6, 'kubernetes', 'cloudnative.music.netease.com/group', 'default'),
(3, 'kubernetes', 'cloudnative.music.netease.com/group', 'default'),
(1, 'kubernetes', 'cloudnative.music.netease.com/group', 'default'),
(8, 'kubernetes', 'cloudnative.music.netease.com/group', 'default'),
(4, 'kubernetes', 'cloudnative.music.netease.com/group', 'default');
