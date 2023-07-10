ALTER TABLE tb_token
ADD COLUMN `ref_id` bigint(20) unsigned null
COMMENT 'id associated to the access token for refresh token';
