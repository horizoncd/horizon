alter table tb_region add column prometheus_url varchar(128) not null default '' comment 'prometheus url' after `ingress_domain`;
