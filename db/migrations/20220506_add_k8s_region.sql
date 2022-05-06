ALTER TABLE tb_region ADD server varchar(256) NULL COMMENT 'k8s server url' AFTER display_name;
ALTER TABLE tb_region ADD certificate text NULL COMMENT 'k8s kube config' AFTER server;
ALTER TABLE tb_region ADD ingress_domain text NULL COMMENT 'k8s ingress domain' AFTER certificate;
