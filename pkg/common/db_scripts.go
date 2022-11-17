package common

/* sql about template */
const (
	TemplateList                  = "select * from tb_template"
	TemplateListByGroup           = "select * from tb_template where group_id = ?"
	TemplateQueryByID             = "select * from tb_template where ID = ?"
	TemplateQueryByName           = "select * from tb_template where name = ?"
	TemplateDelete                = "delete from tb_template where ID = ?"
	TemplateRefCountOfApplication = "select count(*) from tb_template join tb_application " +
		"on tb_template.name = tb_application.template " +
		"where tb_template.id = ? and tb_application.deleted_ts = 0"

	TemplateRefOfApplication = "select tb_application.* from tb_template join tb_application " +
		"on tb_template.name = tb_application.template " +
		"where tb_template.id = ? and tb_application.deleted_ts = 0"

	TemplateRefCountOfCluster = "select count(*) from tb_template join tb_cluster " +
		"on tb_template.name = tb_cluster.template " +
		"where tb_template.id = ? and tb_cluster.deleted_ts = 0"

	TemplateRefOfCluster = "select tb_cluster.* from tb_template join tb_cluster " +
		"on tb_template.name = tb_cluster.template " +
		"where tb_template.id = ? and tb_cluster.deleted_ts = 0"

	TemplateReleaseDelete = "delete from tb_template_release where ID = ?"

	TemplateReleaseQueryByTemplateName = "select tb_template_release.* from tb_template_release join tb_template " +
		"on tb_template_release.template = tb_template.id where tb_template.name = ?"
	TemplateReleaseListByTemplateID           = "select * from tb_template_release where tb_template_release.template = ?"
	TemplateReleaseQueryByTemplateNameAndName = "select * from tb_template_release " +
		"where template_name = ? and name = ?"
	TemplateReleaseQueryByID = "select * from tb_template_release where id = ?"

	TemplateReleaseRefCountOfApplication = "select count(*) from tb_template_release join tb_application " +
		"on tb_template_release.name = tb_application.template_release " +
		"and tb_template_release.template_name = tb_application.template " +
		"where tb_template_release.id = ? and tb_application.deleted_ts = 0"

	TemplateReleaseRefOfApplication = "select tb_application.* from tb_template_release join tb_application " +
		"on tb_template_release.name = tb_application.template_release " +
		"and tb_template_release.template_name = tb_application.template " +
		"where tb_template_release.id = ? and tb_application.deleted_ts = 0"

	TemplateReleaseRefCountOfCluster = "select count(*) from tb_template_release join tb_cluster " +
		"on tb_template_release.name = tb_cluster.template_release " +
		"and tb_template_release.template_name = tb_cluster.template " +
		"where tb_template_release.id = ? and tb_cluster.deleted_ts = 0"

	TemplateReleaseRefOfCluster = "select tb_cluster.* from tb_template_release join tb_cluster " +
		"on tb_template_release.name = tb_cluster.template_release " +
		"and tb_template_release.template_name = tb_cluster.template " +
		"where tb_template_release.id = ? and tb_cluster.deleted_ts = 0"
)

/* sql about user */
const (
	// UserQueryByOIDC ...
	UserQueryByOIDC  = "select * from tb_user where oidc_type = ? and email = ? and user_type = ?"
	UserQueryByEmail = "select * from tb_user where email = ? and user_type = ?"
	UserListByEmail  = "select * from tb_user where email in ? and user_type = ?"
	UserSearch       = "select * from tb_user where user_type = ?" +
		" and (name like ? or full_name like ? or email like ?) limit ? offset ?"
	UserSearchCount = "select count(1) from tb_user where user_type = ?" +
		" and (name like ? or full_name like ? or email like ?)"
	UserGetByID    = "select * from tb_user where id in ?"
	UserDeleteByID = "delete from tb_user where id = ?"
)

/* sql about member */
const (
	MemberQueryByID   = "select * from tb_member where id = ? and deleted_ts = 0"
	MemberSingleQuery = "select * from tb_member where resource_type = ? and  resource_id = ? and member_type= ?" +
		" and membername_id = ? and deleted_ts = 0"
	MemberSingleDelete               = "update tb_member set deleted_ts = ? where ID = ?"
	MemberHardDeleteByResourceTypeID = "delete from tb_member where resource_type = ?" +
		" and resource_id = ?"
	MemberHardDeleteByMemberNameID = "delete from tb_member where membername_id = ?"
	// todo: fix user_type to query condition
	MemberSelectAll = "select m.* from tb_member m join tb_user u on m.membername_id = u.id" +
		" where m.resource_type = ? and m.resource_id = ? and m.deleted_ts = 0"
	// todo: fix user_type to query condition
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
)

/* sql about application */
const (
	ApplicationQueryByIDs             = "select * from tb_application where id in ? and deleted_ts = 0"
	ApplicationQueryByGroupIDs        = "select * from tb_application where group_id in ? and deleted_ts = 0"
	ApplicationQueryByID              = "select * from tb_application where id = ? and deleted_ts = 0"
	ApplicationQueryByName            = "select * from tb_application where name = ? and deleted_ts = 0"
	ApplicationQueryByFuzzily         = "select * from tb_application where name like ? and deleted_ts = 0"
	ApplicationQueryByNamesUnderGroup = "select * from tb_application where group_id = ? and name in ? " +
		"and deleted_ts = 0"
	ApplicationDeleteByID     = "update tb_application set deleted_ts = ?, updated_by = ? where id = ?"
	ApplicationTransferByID   = "update tb_application set group_id = ?, updated_by = ? where id = ?"
	ApplicationCountByGroupID = "select count(1) from tb_application where group_id = ? and deleted_ts = 0"
)

/* sql about environment */
const (
	// EnvironmentListAll ...
	EnvironmentListAll   = "select * from tb_environment where deleted_ts = 0 order by updated_at desc"
	EnvironmentGetByID   = "select * from tb_environment where id = ? and deleted_ts = 0"
	EnvironmentGetByName = "select * from tb_environment where name = ? and deleted_ts = 0"
)

/* sql about environmentRegion */
const (
	EnvironmentListRegion = "select * from tb_environment_region " +
		"where environment_name = ? and deleted_ts = 0"
	EnvironmentListEnabledRegion = "select r.name, r.display_name, r.disabled, er.is_default " +
		"from tb_environment_region er " +
		"join tb_region r on er.region_name = r.name " +
		"where er.environment_name = ? and er.deleted_ts = 0 and r.deleted_ts = 0"
	EnvironmentRegionGet = "select * from tb_environment_region where" +
		" environment_name = ? and region_name = ? and deleted_ts = 0"
	EnvironmentRegionGetByID           = "select * from tb_environment_region where id = ? and deleted_ts = 0"
	EnvironmentRegionListAll           = "select * from tb_environment_region where deleted_ts = 0"
	EnvironmentRegionGetByEnvAndRegion = "select * from tb_environment_region where environment_name = ? and " +
		"region_name = ? and deleted_ts = 0"
	EnvironmentRegionGetDefaultByEnv = "select * from tb_environment_region where environment_name = ? and " +
		"is_default = 1 and deleted_ts = 0"
	EnvironmentRegionsGetDefault = "select * from tb_environment_region where " +
		"is_default = 1 and deleted_ts = 0"
	EnvironmentRegionSetDefaultByID   = "update tb_environment_region set is_default = 1 where id = ?"
	EnvironmentRegionUnsetDefaultByID = "update tb_environment_region set is_default = 0 where id = ?"
)

/* sql about region */
const (
	// RegionListAll ...
	RegionListAll    = "select * from tb_region where deleted_ts = 0 order by updated_at desc"
	RegionGetByName  = "select * from tb_region where name = ? and deleted_ts = 0"
	RegionGetByID    = "select * from tb_region where id = ? and deleted_ts = 0"
	RegionListByTags = "select r.name, r.display_name, r.disabled from tb_region r " +
		"join tb_tag tg on r.id = tg.resource_id " +
		"where tg.resource_type = ? and r.deleted_ts = 0 " +
		"and %s group by r.id having count(r.id) = ?"
)

/* sql about cluster */
const (
	ClusterCountByRegionName  = "select count(1) from tb_cluster where region_name = ? and deleted_ts = 0"
	ClusterQueryByID          = "select * from tb_cluster where id = ? and deleted_ts = 0"
	ClusterDeleteByID         = "update tb_cluster set deleted_ts = ?, updated_by = ? where id = ?"
	ClusterQueryByName        = "select * from tb_cluster where name = ? and deleted_ts = 0"
	ClusterQueryByClusterName = "select * from tb_cluster where name = ? and deleted_ts = 0"
)

/* sql about pipelinerun */
const (
	PipelinerunGetByID                       = "select * from tb_pipelinerun where id = ?"
	PipelinerunDeleteByID                    = "delete from tb_pipelinerun where id = ?"
	PipelinerunDeleteByClusterID             = "delete from tb_pipelinerun where cluster_id = ?"
	PipelinerunUpdateConfigCommitByID        = "update tb_pipelinerun set config_commit = ? where id = ?"
	PipelinerunGetLatestByClusterIDAndAction = "select * from tb_pipelinerun where cluster_id = ? " +
		"and action = ? order by id desc limit 1"
	PipelinerunGetLatestByClusterIDAndActionAndStatus = "select * from tb_pipelinerun where cluster_id = ? " +
		"and action = ? and status = ? order by id desc limit 1"
	PipelinerunGetLatestSuccessByClusterID = "select * from tb_pipelinerun where cluster_id = ? and status = 'ok' and " +
		"git_commit != '' order by updated_at desc limit 1"

	PipelinerunUpdateStatusByID = "update tb_pipelinerun set status = ? where id = ?"
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
	// TagListByResourceTypeID ...
	TagListByResourceTypeID = "select * from tb_tag where resource_type = ?" +
		" and resource_id = ? order by id"
	TagListByResourceTypeIDs = "select * from tb_tag where resource_type = ?" +
		" and resource_id in ? order by id"
	TagListDistinctByResourceTypeIDs = "select distinct tag_key, tag_value from tb_tag where resource_type = ?" +
		" and resource_id in ?"
	TagDeleteAllByResourceTypeID = "delete from tb_tag where resource_type = ?" +
		" and resource_id = ?"
	TagDeleteByResourceTypeIDAndKeys = "delete from tb_tag where resource_type = ?" +
		" and resource_id = ? and `tag_key` not in ?"
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
	ApplicationRegionListByEnvApplicationID = "select * from tb_application_region where environment_name = ? " +
		"and application_id = ?"
	ApplicationRegionListByApplicationID      = "select * from tb_application_region where application_id = ?"
	ApplicationRegionDeleteAllByApplicationID = "delete from tb_application_region where application_id = ?"
)

/* sql about token*/
const (
	DeleteByCode     = "delete  from tb_token where code = ?"
	DeleteTokenByID  = "delete from tb_token where id = ?"
	TokenGetByCode   = "select * from tb_token where code = ?"
	DeleteByClientID = "delete from tb_token where client_id = ?"
)

/* sql about oauth app*/
const (
	GetOauthAppByClientID        = "select * from tb_oauth_app where  client_id = ?"
	DeleteOauthAppByClientID     = "delete from tb_oauth_app where client_id = ?"
	SelectOauthAppByOwner        = "select * from tb_oauth_app  where owner_type = ? and owner_id = ?"
	DeleteClientSecret           = "delete from tb_oauth_client_secret where  client_id = ? and id = ?"
	DeleteClientSecretByClientID = "delete from tb_oauth_client_secret where client_id = ?"
	ClientSecretSelectAll        = "select * from tb_oauth_client_secret where client_id = ?"
)

/* sql about pipeline*/
const (
	PipelineDeleteByCluster = "delete from tb_pipeline where cluster= ?"
	TaskDeleteByCluster     = "delete from tb_task where cluster= ?"
	StepDeleteByCluster     = "delete from tb_step where cluster= ?"
)
