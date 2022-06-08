-- set region to be default
ALTER TABLE horizon.tb_environment_region ADD is_default tinyint(1) NOT NULL DEFAULT 0 COMMENT '0 means not default region, 1 means default region' AFTER region_name;

-- this field is not used anymore
Alter TABLE horizon.tb_region DROP k8s_cluster_id;

-- all fields has been moved to tb_region, this table can be dropped
DROP TABLE horizon.tb_k8s_cluster;

-- add name field to harbor
ALTER TABLE horizon.tb_harbor ADD name varchar(128) NOT NULL DEFAULT '' COMMENT 'name of the harbor registry' AFTER id;
