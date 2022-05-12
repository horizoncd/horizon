-- merge k8s_cluster to region
ALTER TABLE tb_region ADD server varchar(256) NULL COMMENT 'k8s server url' AFTER display_name;
ALTER TABLE tb_region ADD certificate text NULL COMMENT 'k8s kube config' AFTER server;
ALTER TABLE tb_region ADD ingress_domain text NULL COMMENT 'k8s ingress domain' AFTER certificate;

-- add default_region to environment
ALTER TABLE tb_environment ADD default_region varchar(128) NULL COMMENT 'default region of the environment' AFTER display_name;
