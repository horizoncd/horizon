package common

/* sql about template */
const (
	TemplateQuery                      = "select * from tb_template where deleted_ts = 0"
	TemplateReleaseQueryByTemplateName = "select * from tb_template_release " +
		"where template_name = ? and deleted_ts = 0"
	TemplateReleaseQueryByTemplateNameAndName = "select * from tb_template_release " +
		"where template_name = ? and name = ? and deleted_ts = 0"
)

/* sql about user */
const (
	// UserQueryByOIDC ...
	UserQueryByOIDC  = "select * from tb_user where oidc_type = ? and email = ?"
	UserQueryByEmail = "select * from tb_user where email = ? "
	UserListByEmail  = "select * from tb_user where email in ? "
	UserSearch       = "select * from tb_user where name like ? or full_name like ? or email like ? limit ? offset ?"
	UserSearchCount  = "select count(1) from tb_user where name like ? or full_name like ? or email like ?"
	UserGetByID      = "select * from tb_user where id in ?"
)

/* sql about member */
const (
	MemberQueryByID   = "select * from tb_member where id = ? and deleted_ts = 0"
	MemberSingleQuery = "select * from tb_member where resource_type = ? and  resource_id = ? and member_type= ?" +
		" and membername_id = ? and deleted_ts = 0"
	MemberSingleDelete       = "update tb_member set deleted_ts = ? where ID = ?"
	MemberSelectAll          = "select * from tb_member where resource_type = ? and resource_id = ? and deleted_ts = 0"
	MemberSelectByUserEmails = "select tb_member.* from tb_member join tb_user on tb_member.membername_id = tb_user.id" +
		" where tb_member.resource_type = ? and tb_member.resource_id = ? and tb_user.email in ?" +
		" and tb_member.member_type = 0 and tb_member.deleted_ts = 0 and tb_user.deleted_ts = 0"
	MemberListResource = "select resource_id from tb_member where resource_type = ? and" +
		" membername_id = ? and deleted_ts = 0"
)

/* sql about group */
const (
	GroupQueryByParentIDAndName = "select * from tb_group where parent_id = ? and name = ? and deleted_ts = 0"
	GroupQueryByParentIDAndPath = "select * from tb_group where parent_id = ? and path = ? and deleted_ts = 0"
	GroupDelete                 = "update tb_group set deleted_ts = ?, updated_by = ? where id = ?"
	GroupUpdateBasic            = "update tb_group set name = ?, path = ?, description = ?, visibility_level = ?, " +
		"updated_by = ? where id = ?"
	GroupUpdateParentID       = "update tb_group set parent_id = ?, updated_by = ? where id = ?"
	GroupQueryByID            = "select * from tb_group where id = ? and deleted_ts = 0"
	GroupQueryByIDs           = "select * from tb_group where id in ? and deleted_ts = 0"
	GroupQueryByPaths         = "select * from tb_group where path in ? and deleted_ts = 0"
	GroupQueryByNameFuzzily   = "select * from tb_group where name like ? and deleted_ts = 0"
	GroupQueryByIDNameFuzzily = "select * from tb_group " +
		"where traversal_ids like ? and name like ? and deleted_ts = 0"
	GroupAll                      = "select * from tb_group where deleted_ts = 0"
	GroupUpdateTraversalIDs       = "update tb_group set traversal_ids = ?, updated_by = ? where id = ? and deleted_ts = 0"
	GroupCountByParentID          = "select count(1) from tb_group where parent_id = ? and deleted_ts = 0"
	GroupUpdateTraversalIDsPrefix = "update tb_group set traversal_ids = replace(traversal_ids, ?, ?), updated_by = ? " +
		"where traversal_ids like ? and deleted_ts = 0"
	GroupQueryByNameOrPathUnderParent = "select * from tb_group where parent_id = ? " +
		"and (name = ? or path = ?) and deleted_ts = 0"
	GroupQueryGroupChildren = "" +
		"select * from (select g.id, g.name, g.path, description, updated_at, 'group' as type from tb_group g " +
		"where g.parent_id=? and g.deleted_ts = 0 " +
		"union " +
		"select a.id, a.name, a.name as path, description, updated_at, 'application' as type from tb_application a " +
		"where a.group_id=? and a.deleted_ts = 0) ga " +
		"order by ga.type desc,ga.updated_at desc limit ? offset ?"
	GroupQueryGroupChildrenCount = "" +
		"select count(1) from (select g.id, g.name, g.path, description, updated_at, 'group' as type from tb_group g " +
		"where g.parent_id=? and g.deleted_ts = 0 " +
		"union " +
		"select a.id, a.name, a.name as path, description, updated_at, 'application' as type from tb_application a " +
		"where a.group_id=? and a.deleted_ts = 0) ga"
	GroupQueryByTraversalID = "select * from tb_group where %s and deleted_ts = 0"
)

/* sql about application */
const (
	ApplicationQueryByIDs                  = "select * from tb_application where id in ? and deleted_ts = 0"
	ApplicationQueryByGroupIDs             = "select * from tb_application where group_id in ? and deleted_ts = 0"
	ApplicationQueryByID                   = "select * from tb_application where id = ? and deleted_ts = 0"
	ApplicationQueryByName                 = "select * from tb_application where name = ? and deleted_ts = 0"
	ApplicationQueryByFuzzily              = "select * from tb_application where name like ? and deleted_ts = 0"
	ApplicationQueryByFuzzilyCount         = "select count(1) from tb_application where name like ? and deleted_ts = 0"
	ApplicationQueryByFuzzilyAndPagination = "select * from tb_application where name like ? and deleted_ts = 0 " +
		"order by updated_at desc limit ? offset ?"
	ApplicationQueryByNamesUnderGroup = "select * from tb_application where group_id = ? and name in ? " +
		"and deleted_ts = 0"
	ApplicationDeleteByID                = "update tb_application set deleted_ts = ?, updated_by = ? where id = ?"
	ApplicationTransferByID              = "update tb_application set group_id = ?, updated_by = ? where id = ?"
	ApplicationCountByGroupID            = "select count(1) from tb_application where group_id = ? and deleted_ts = 0"
	ApplicationQueryByUserAndNameFuzzily = "select * from ( " +
		"select a.* from tb_application a join tb_member m on m.resource_id = a.id " +
		"where m.resource_type = 'applications' and m.member_type = '0' and m.membername_id = ? " +
		"and a.name like ? and a.deleted_ts = 0 and m.deleted_ts = 0 " +
		"union " +
		"select a.* from tb_application a where a.group_id in ? " +
		"and a.name like ? and a.deleted_ts = 0" +
		") da order by updated_at desc limit ? offset ?"
	ApplicationCountByUserAndNameFuzzily = "select count(1) from ( " +
		"select a.* from tb_application a join tb_member m on m.resource_id = a.id " +
		"where m.resource_type = 'applications' and m.member_type = '0' and m.membername_id = ? " +
		"and a.name like ? and a.deleted_ts = 0 and m.deleted_ts = 0 " +
		"union " +
		"select a.* from tb_application a where a.group_id in ? " +
		"and a.name like ? and a.deleted_ts = 0) da"
)

/* sql about harbor */
const (
	HarborListAll = "select * from tb_harbor where deleted_ts = 0"
	HarborGetByID = "select * from tb_harbor where id = ? and deleted_ts = 0"
)

/* sql about environment */
const (
	// EnvironmentListAll ...
	EnvironmentListAll    = "select * from tb_environment where deleted_ts = 0"
	EnvironmentListRegion = "select region_name from tb_environment_region " +
		"where environment_name = ? and disabled = 0 and deleted_ts = 0"
	EnvironmentRegionGet = "select * from tb_environment_region where" +
		" environment_name = ? and region_name = ? and deleted_ts = 0"
	EnvironmentRegionGetByID = "select * from tb_environment_region where id = ? and deleted_ts = 0"
)

/* sql about region */
const (
	// RegionListAll ...
	RegionListAll     = "select * from tb_region where deleted_ts = 0"
	RegionGetByName   = "select * from tb_region where name = ? and deleted_ts = 0"
	RegionListByNames = "select * from tb_region where name in ? and deleted_ts = 0"
)

/* sql about cluster */
const (
	ClusterQueryByID           = "select * from tb_cluster where id = ? and deleted_ts = 0"
	ClusterDeleteByID          = "update tb_cluster set deleted_ts = ?, updated_by = ? where id = ?"
	ClusterQueryByName         = "select * from tb_cluster where name = ? and deleted_ts = 0"
	ClusterListByApplicationID = "select * from tb_cluster where application_id = ? and deleted_ts = 0"
	ClusterQueryByApplication  = "select c.*, er.environment_name, er.region_name, " +
		"r.display_name as region_display_name from tb_cluster c " +
		"join tb_environment_region er on c.environment_region_id = er.id " +
		"join tb_region r on r.name = er.region_name " +
		"where c.application_id = ? " +
		"and c.name like ? and c.deleted_ts = 0 limit ? offset ?"
	ClusterCountByApplication = "select count(1) from tb_cluster c " +
		"join tb_environment_region er on c.environment_region_id = er.id " +
		"join tb_region r on r.name = er.region_name " +
		"where c.application_id = ? " +
		"and c.name like ? and c.deleted_ts = 0"
	ClusterQueryByApplicationAndEnvs = "select c.*, er.environment_name, er.region_name, " +
		"r.display_name as region_display_name from tb_cluster c " +
		"join tb_environment_region er on c.environment_region_id = er.id " +
		"join tb_region r on r.name = er.region_name " +
		"where c.application_id = ? and er.environment_name in ? " +
		"and c.name like ? and c.deleted_ts = 0 limit ? offset ?"
	ClusterCountByApplicationAndEnvs = "select count(1) from tb_cluster c " +
		"join tb_environment_region er on c.environment_region_id = er.id " +
		"join tb_region r on r.name = er.region_name " +
		"where c.application_id = ? and er.environment_name in ? " +
		"and c.name like ? and c.deleted_ts = 0"
	ClusterQueryByNameFuzzily = "select c.*, er.environment_name, er.region_name, " +
		"r.display_name as region_display_name from tb_cluster c " +
		"join tb_environment_region er on c.environment_region_id = er.id " +
		"join tb_region r on r.name = er.region_name " +
		"where %s c.name like ? and c.deleted_ts = 0 order by updated_at desc limit ? offset ?"
	ClusterCountByNameFuzzily = "select count(1) from tb_cluster c " +
		"join tb_environment_region er on c.environment_region_id = er.id " +
		"join tb_region r on r.name = er.region_name " +
		"where %s c.name like ? and c.deleted_ts = 0"
	ClusterQueryByEnvNameFuzzily = "select c.*, er.environment_name, er.region_name, " +
		"r.display_name as region_display_name from tb_cluster c " +
		"join tb_environment_region er on c.environment_region_id = er.id " +
		"join tb_region r on r.name = er.region_name " +
		"where %s er.environment_name = ? and c.name like ? and c.deleted_ts = 0 " +
		"order by updated_at desc limit ? offset ?"
	ClusterCountByEnvNameFuzzily = "select count(1) from tb_cluster c " +
		"join tb_environment_region er on c.environment_region_id = er.id " +
		"join tb_region r on r.name = er.region_name " +
		"where %s er.environment_name = ? and c.name like ? and c.deleted_ts = 0"
	ClusterQueryByClusterName        = "select * from tb_cluster where name = ? and deleted_ts = 0"
	ClusterQueryByUserAndNameFuzzily = "select * from (" +
		"select c.*, er.environment_name, er.region_name, r.display_name as region_display_name " +
		"from tb_cluster c join tb_member m on m.resource_id = c.id " +
		"join tb_environment_region er on c.environment_region_id = er.id  " +
		"join tb_region r on r.name = er.region_name " +
		"where %s m.resource_type = 'clusters' and m.member_type = '0' and m.membername_id = ? and c.name like ? " +
		"and c.deleted_ts = 0 and m.deleted_ts = 0 " +
		"union " +
		"select c.*, er.environment_name, er.region_name, r.display_name as region_display_name " +
		"from tb_cluster c " +
		"join tb_environment_region er on c.environment_region_id = er.id  " +
		"join tb_region r on r.name = er.region_name " +
		"where %s c.application_id in ? and c.name like ? and c.deleted_ts = 0) " +
		"dc order by updated_at desc limit ? offset ?"
	ClusterCountByUserAndNameFuzzily = "select count(1) from (" +
		"select c.*, er.environment_name, er.region_name, r.display_name as region_display_name " +
		"from tb_cluster c join tb_member m on m.resource_id = c.id " +
		"join tb_environment_region er on c.environment_region_id = er.id  " +
		"join tb_region r on r.name = er.region_name " +
		"where %s m.resource_type = 'clusters' and m.member_type = '0' and m.membername_id = ? and c.name like ? " +
		"and c.deleted_ts = 0 and m.deleted_ts = 0 " +
		"union " +
		"select c.*, er.environment_name, er.region_name, r.display_name as region_display_name " +
		"from tb_cluster c " +
		"join tb_environment_region er on c.environment_region_id = er.id  " +
		"join tb_region r on r.name = er.region_name " +
		"where %s c.application_id in ? and c.name like ? and c.deleted_ts = 0) dc"
	ClusterQueryByUserAndEnvAndNameFuzzily = "select * from (" +
		"select c.*, er.environment_name, er.region_name, r.display_name as region_display_name " +
		"from tb_cluster c join tb_member m on m.resource_id = c.id " +
		"join tb_environment_region er on c.environment_region_id = er.id  " +
		"join tb_region r on r.name = er.region_name " +
		"where %s m.resource_type = 'clusters' and m.member_type = '0' " +
		"and m.membername_id = ? and er.environment_name = ? and c.name like ? " +
		"and c.deleted_ts = 0 and m.deleted_ts = 0 " +
		"union " +
		"select c.*, er.environment_name, er.region_name, r.display_name as region_display_name " +
		"from tb_cluster c " +
		"join tb_environment_region er on c.environment_region_id = er.id  " +
		"join tb_region r on r.name = er.region_name " +
		"where %s c.application_id in ? and er.environment_name = ? and c.name like ? and c.deleted_ts = 0) " +
		"dc order by updated_at desc limit ? offset ?"
	ClusterCountByUserAndEnvAndNameFuzzily = "select count(1) from (" +
		"select c.*, er.environment_name, er.region_name, r.display_name as region_display_name " +
		"from tb_cluster c join tb_member m on m.resource_id = c.id " +
		"join tb_environment_region er on c.environment_region_id = er.id  " +
		"join tb_region r on r.name = er.region_name " +
		"where %s m.resource_type = 'clusters' and m.member_type = '0' " +
		"and m.membername_id = ? and er.environment_name = ? and c.name like ? " +
		"and c.deleted_ts = 0 and m.deleted_ts = 0 " +
		"union " +
		"select c.*, er.environment_name, er.region_name, r.display_name as region_display_name " +
		"from tb_cluster c " +
		"join tb_environment_region er on c.environment_region_id = er.id  " +
		"join tb_region r on r.name = er.region_name " +
		"where %s c.application_id in ? and er.environment_name = ? and c.name like ? and c.deleted_ts = 0) dc"
)

/* sql about pipelinerun */
const (
	PipelinerunGetByID                       = "select * from tb_pipelinerun where id = ?"
	PipelinerunDeleteByID                    = "delete from tb_pipelinerun where id = ?"
	PipelinerunUpdateConfigCommitByID        = "update tb_pipelinerun set config_commit = ? where id = ?"
	PipelinerunGetLatestByClusterIDAndAction = "select * from tb_pipelinerun where cluster_id = ? " +
		"and action = ? order by id desc limit 1"
	PipelinerunGetLatestByClusterIDAndActionAndStatus = "select * from tb_pipelinerun where cluster_id = ? " +
		"and action = ? and status = ? order by id desc limit 1"
	PipelinerunGetLatestSuccessByClusterID = "select * from tb_pipelinerun where cluster_id = ? and status = 'ok' and " +
		"git_commit != '' order by updated_at desc limit 1"
	PipelinerunUpdateResultByID = "update tb_pipelinerun set status = ?, s3_bucket = ?, log_object = ?, " +
		"pr_object = ?, started_at = ?, finished_at = ? where id = ?"

	PipelinerunGetByClusterID = "select * from tb_pipelinerun where cluster_id = ?" +
		" order by created_at desc limit ? offset ?"

	PipelinerunGetByClusterIDTotalCount = "select count(1) from tb_pipelinerun where cluster_id = ?"

	PipelinerunCanRollbackGetByClusterID = "select * from tb_pipelinerun where cluster_id = ?" +
		" and action != 'restart' and status = 'ok' order by created_at desc limit ? offset ?"

	PipelinerunCanRollbackGetByClusterIDTotalCount = "select count(1) - 1 from tb_pipelinerun " +
		"where cluster_id = ? and action != 'restart' and status = 'ok' "

	PipelinerunGetFirstCanRollbackByClusterID = "select * from tb_pipelinerun where cluster_id = ?" +
		" and action != 'restart' and status = 'ok' order by created_at desc limit 1 offset 0"
)

/* sql about cluster tag */
const (
	// ClusterTagListByClusterID ...
	ClusterTagListByClusterID          = "select * from tb_cluster_tag where cluster_id = ? order by id"
	ClusterTagDeleteAllByClusterID     = "delete from tb_cluster_tag where cluster_id = ?"
	ClusterTagDeleteByClusterIDAndKeys = "delete from tb_cluster_tag where cluster_id = ? and `tag_key` not in ?"
)

/* sql about cluster template tag */
const (
	ClusterTemplateSchemaTagListByClusterID = "select * from tb_cluster_template_schema_tag where cluster_id = ? " +
		"order by id"
	ClusterTemplateSchemaTagDeleteAllByClusterID     = "delete from tb_cluster_template_schema_tag where cluster_id = ?"
	ClusterTemplateSchemaTagDeleteByClusterIDAndKeys = "delete from tb_cluster_template_schema_tag where cluster_id = ?" +
		" and `tag_key` not in ?"
)

/* sql about application region */
const (
	ApplicationRegionListByApplicationID      = "select * from tb_application_region where application_id = ?"
	ApplicationRegionDeleteAllByApplicationID = "delete from tb_application_region where application_id = ?"
)

/* sql about token*/
const (
	DeleteByCode     = "delete from tb_token where code = ?"
	TokenGetByCode   = "select * from tb_token where code = ?"
	DeleteByClientID = "delete * from tb_token where client_id = ?"
)

/* sql about oauth app*/
const (
	GetOauthAppByClientID        = "select * from tb_oauth_app where  client_id = ?"
	DeleteOauthAppByClientID     = "delete * from tb_oauth_app where client_id = ?"
	DeleteClientSecret           = "delete * from tb_oauth_client_secret where  client_id =? and id = ?"
	DeleteClientSecretByClientID = "delete * from tb_oauth_client_secret where client_id = ?"
	ClientSecretSelectAll        = "select * from tb_oauth_client_secret where client_id = ?"
)
