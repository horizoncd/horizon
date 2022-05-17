-- add default_region to environment
ALTER TABLE tb_environment_region ADD is_default tinyint(1) NOT NULL DEFAULT 0 COMMENT '0 means not default region, 1 means default region' AFTER region_name;
Alter TABLE tb_region DROP k8s_cluster_id;
DROP TABLE tb_k8s_cluster;
