package common

/* sql about template */
const (
	TemplateQuery                      = "select * from template where deleted_at is null"
	TemplateReleaseQueryByTemplateName = "select * from template_release " +
		"where template_name = ? and deleted_at is null"
	TemplateReleaseQueryByTemplateNameAndName = "select * from template_release " +
		"where template_name = ? and name = ? and deleted_at is null"
)

/* sql about user */
const (
	// UserQueryByOIDC ...
	UserQueryByOIDC  = "select * from user where oidc_type = ? and email = ?"
	UserQueryByEmail = "select * from user where email = ? "
	UserSearch       = "select * from user where name like ? or full_name like ? or email like ? limit ? offset ?"
	UserSearchCount  = "select count(1) from user where name like ? or full_name like ? or email like ?"
	UserGetByID      = "select * from user where id in ?"
)

/* sql about member */
const (
	MemberQueryByID   = "select * from member where id = ? and deleted_at is null"
	MemberSingleQuery = "select * from member where resource_type = ? and  resource_id = ? and member_type= ?" +
		" and membername_id = ? and deleted_at is null"
	MemberSingleDelete = "update member set deleted_at = CURRENT_TIMESTAMP where ID = ?"
	MemberSelectAll    = "select * from member where resource_type = ? and resource_id = ? and deleted_at is null"
)

/* sql about group */
const (
	GroupQueryByParentIDAndName = "select * from `group` where parent_id = ? and name = ? and deleted_at is null"
	GroupQueryByParentIDAndPath = "select * from `group` where parent_id = ? and path = ? and deleted_at is null"
	GroupDelete                 = "update `group` set deleted_at = CURRENT_TIMESTAMP where id = ?"
	GroupUpdateBasic            = "update `group` set name = ?, path = ?, description = ?, visibility_level = ?, " +
		"updated_by = ? where id = ?"
	GroupUpdateParentID       = "update `group` set parent_id = ?, updated_by = ? where id = ?"
	GroupQueryByID            = "select * from `group` where id = ? and deleted_at is null"
	GroupQueryByIDs           = "select * from `group` where id in ? and deleted_at is null"
	GroupQueryByPaths         = "select * from `group` where path in ? and deleted_at is null"
	GroupQueryByNameFuzzily   = "select * from `group` where name like ? and deleted_at is null"
	GroupQueryByIDNameFuzzily = "select * from `group` " +
		"where traversal_ids like ? and name like ? and deleted_at is null"
	GroupUpdateTraversalIDs       = "update `group` set traversal_ids = ? where id = ? and deleted_at is null"
	GroupCountByParentID          = "select count(1) from `group` where parent_id = ? and deleted_at is null"
	GroupUpdateTraversalIDsPrefix = "update `group` set traversal_ids = replace(traversal_ids, ?, ?), updated_by = ? " +
		"where traversal_ids like ? and deleted_at is null"
	GroupQueryByNameOrPathUnderParent = "select * from `group` where parent_id = ? " +
		"and (name = ? or path = ?) and deleted_at is null"
	GroupQueryGroupChildren = "" +
		"select * from (select g.id, g.name, g.path, description, updated_at, 'group' as type from `group` g " +
		"where g.parent_id=? and g.deleted_at is null " +
		"union " +
		"select a.id, a.name, a.name as path, description, updated_at, 'application' as type from `application` a " +
		"where a.group_id=? and a.deleted_at is null) ga " +
		"order by ga.type desc,ga.updated_at desc limit ? offset ?"
	GroupQueryGroupChildrenCount = "" +
		"select count(1) from (select g.id, g.name, g.path, description, updated_at, 'group' as type from `group` g " +
		"where g.parent_id=? and g.deleted_at is null " +
		"union " +
		"select a.id, a.name, a.name as path, description, updated_at, 'application' as type from `application` a " +
		"where a.group_id=? and a.deleted_at is null) ga"
)

/* sql about application */
const (
	ApplicationQueryByIDs                  = "select * from application where id in ? and deleted_at is null"
	ApplicationQueryByID                   = "select * from application where id = ? and deleted_at is null"
	ApplicationQueryByName                 = "select * from application where name = ? and deleted_at is null"
	ApplicationQueryByFuzzily              = "select * from application where name like ? and deleted_at is null"
	ApplicationQueryByFuzzilyCount         = "select count(1) from application where name like ? and deleted_at is null"
	ApplicationQueryByFuzzilyAndPagination = "select * from application where name like ? and deleted_at is null " +
		"order by name desc limit ? offset ?"
	ApplicationQueryByNamesUnderGroup = "select * from application where group_id = ? and name in ? " +
		"and deleted_at is null"
	ApplicationDeleteByID     = "update application set deleted_at = CURRENT_TIMESTAMP where id = ?"
	ApplicationCountByGroupID = "select count(1) from application where group_id = ? and deleted_at is null"
)

/* sql about k8sCluster */
const (
	// K8SClusterListAll ...
	K8SClusterListAll     = "select * from k8s_cluster where deleted_at is null"
	K8SClusterGetByServer = "select * from k8s_cluster where server = ? and deleted_at is null"
)

/* sql about harbor */
const (
	HarborListAll = "select * from harbor where deleted_at is null"
	HarborGetByID = "select * from harbor where id = ? and deleted_at is null"
)

/* sql about environment */
const (
	// EnvironmentListAll ...
	EnvironmentListAll    = "select * from environment where deleted_at is null"
	EnvironmentListRegion = "select region_name from environment_region where environment_name = ? and deleted_at is null"
	EnvironmentRegionGet  = "select * from environment_region where" +
		" environment_name = ? and region_name = ? and deleted_at is null"
	EnvironmentRegionGetByID = "select * from environment_region where id = ? and deleted_at is null"
)

/* sql about region */
const (
	// RegionListAll ...
	RegionListAll     = "select * from region where deleted_at is null"
	RegionGetByName   = "select * from region where name = ? and deleted_at is null"
	RegionListByNames = "select * from region where name in ? and deleted_at is null"
)

/* sql about cluster */
const (
	ClusterQueryByID                = "select * from cluster where id = ? and deleted_at is null"
	ClusterQueryByName              = "select * from cluster where name = ? and deleted_at is null"
	ClusterQueryByApplicationAndEnv = "select c.*, er.environment_name, er.region_name, " +
		"r.display_name as region_display_name from cluster c " +
		"join environment_region er on c.environment_region_id = er.id " +
		"join region r on r.name = er.region_name " +
		"where c.application_id = ? and er.environment_name = ? " +
		"and c.name like ? and c.deleted_at is null limit ? offset ?"
	ClusterCountByApplicationAndEnv = "select count(1) from cluster c " +
		"join environment_region er on c.environment_region_id = er.id " +
		"join region r on r.name = er.region_name " +
		"where c.application_id = ? and er.environment_name = ? " +
		"and c.name like ? and c.deleted_at is null"
	ClusterQueryByNameFuzzily = "select c.*, er.environment_name, er.region_name, " +
		"r.display_name as region_display_name from cluster c " +
		"join environment_region er on c.environment_region_id = er.id " +
		"join region r on r.name = er.region_name " +
		"where c.name like ? and c.deleted_at is null limit ? offset ?"
	ClusterCountByNameFuzzily = "select count(1) from cluster c " +
		"join environment_region er on c.environment_region_id = er.id " +
		"join region r on r.name = er.region_name " +
		"where c.name like ? and c.deleted_at is null"
	ClusterQueryByEnvNameFuzzily = "select c.*, er.environment_name, er.region_name, " +
		"r.display_name as region_display_name from cluster c " +
		"join environment_region er on c.environment_region_id = er.id " +
		"join region r on r.name = er.region_name " +
		"where er.environment_name = ? and c.name like ? and c.deleted_at is null limit ? offset ?"
	ClusterCountByEnvNameFuzzily = "select count(1) from cluster c " +
		"join environment_region er on c.environment_region_id = er.id " +
		"join region r on r.name = er.region_name " +
		"where er.environment_name = ? and c.name like ? and c.deleted_at is null"
	ClusterQueryByClusterName = "select * from cluster where name = ? and deleted_at is null"
)

/* sql about pipelinerun */
const (
	PipelinerunGetByID                       = "select * from pipelinerun where id = ?"
	PipelinerunDeleteByID                    = "delete from pipelinerun where id = ?"
	PipelinerunUpdateConfigCommitByID        = "update pipelinerun set config_commit = ? where id = ?"
	PipelinerunGetLatestByClusterIDAndAction = "select * from pipelinerun where cluster_id = ? " +
		"and action = ? order by id desc limit 1"
)
