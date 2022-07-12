ALTER TABLE horizon.tb_pipelinerun ADD git_ref varchar(128) NULL after git_branch;
ALTER TABLE horizon.tb_cluster ADD git_ref varchar(128) NULL after git_branch;
ALTER TABLE horizon.tb_application ADD git_ref varchar(128) NULL after git_branch;

ALTER TABLE horizon.tb_pipelinerun ADD git_ref_type varchar(64) NULL after git_ref;
ALTER TABLE horizon.tb_cluster ADD git_ref_type varchar(64) NULL after git_ref;
ALTER TABLE horizon.tb_application ADD git_ref_type varchar(64) NULL after git_ref;

UPDATE horizon.tb_cluster set git_ref_type = 'branch';
UPDATE horizon.tb_application set git_ref_type = 'branch';
UPDATE horizon.tb_pipelinerun set git_ref_type = 'branch';

UPDATE horizon.tb_cluster set git_ref = git_branch;
UPDATE horizon.tb_application set git_ref = git_branch;
UPDATE horizon.tb_pipelinerun set git_ref = git_branch;