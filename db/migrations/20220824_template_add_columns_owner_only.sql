alter table `tb_template` add column `only_owner` boolean not null default false comment 'only owner could access';
alter table `tb_template_release` add column `only_owner` boolean not null default false comment 'only owner could access';
