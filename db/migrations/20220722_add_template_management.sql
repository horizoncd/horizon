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

alter table tb_template add column repository varchar(256) not null default '' after `description`;
alter table tb_template add column group_id bigint(20) unsigned not null default 0 after `repository`;
alter table tb_template add column chart_name varchar(256) default '' after `name`;
alter table tb_template add column only_admin bool not null default false after `group_id`;

alter table tb_template_release add column template bigint(20) unsigned not null default 0 after `template_name`;
alter table tb_template_release add column chart_name varchar(256) not null default '' after `name`;
alter table tb_template_release add column only_admin bool not null default false after recommended;
alter table tb_template_release add column chart_version varchar(256) not null default '' comment 'chart version on template repository' after `chart_name`;
alter table tb_template_release add column sync_status varchar(64) not null default 'status_unknown' comment 'shows sync status' after `chart_version`;
alter table tb_template_release add column failed_reason varchar(2048) not null default '' comment 'failed reason at last time' after `sync_status`;
alter table tb_template_release add column commit_id varchar(256) not null default '' comment 'commit id at last sync' after `failed_reason`;
alter table tb_template_release add column last_sync_at datetime not null default current_timestamp after sync_status;

update tb_template set chart_name = `name` where chart_name = '';
update tb_template_release set `chart_version` = `name` where `chart_version` = '';
update tb_template_release a, tb_template b set a.template = b.id where a.template_name = b.name;
update tb_template_release a, tb_template b set a.chart_name = b.chart_name where a.template = b.id;