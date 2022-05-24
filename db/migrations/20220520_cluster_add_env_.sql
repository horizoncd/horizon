alter table tb_cluster
    add environment_name varchar(128) default '' not null after name;
alter table tb_cluster
    add region_name varchar(128) default '' not null after environment_name;
