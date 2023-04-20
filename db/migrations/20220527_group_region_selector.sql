-- Copyright © 2023 Horizoncd.
--
-- Licensed under the Apache License, Version 2.0 (the "License");
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
--     http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an "AS IS" BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.

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
