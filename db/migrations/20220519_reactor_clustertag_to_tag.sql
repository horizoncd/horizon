-- add resource_type field to tb_tag, and update to "cluster"
ALTER TABLE `horizon`.`tb_cluster_tag` ADD resource_type varchar(64) NOT NULL DEFAULT "" COMMENT 'resource type' AFTER cluster_id;
update `horizon`.`tb_cluster_tag` set resource_type="clusters";

-- add new index and remove old index
CREATE UNIQUE INDEX idx_resource_key ON `horizon`.`tb_cluster_tag` (`resource_type`, `cluster_id`, `tag_key`);
DROP INDEX idx_cluster_id_key ON `horizon`.`tb_cluster_tag`;

-- rename cluster_id field to resource_id
ALTER TABLE `horizon`.`tb_cluster_tag` CHANGE `cluster_id` `resource_id` bigint(20) unsigned NOT NULL COMMENT 'resource id';

-- rename tb_cluster_tag to tb_tag
alter table `horizon`.`tb_cluster_tag` rename to `horizon`.`tb_tag`;