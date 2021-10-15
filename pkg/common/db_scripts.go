package common

/* sql about gitlab */
const (
	GitlabQuery       = "select * from gitlab where deleted_at is null"
	GitlabQueryByName = "select * from gitlab where name = ? and deleted_at is null"
)

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
	UserQueryByOIDC = "select * from user where oidc_id = ? and oidc_type = ?"
	UserSearch      = "select * from user where name like ? or full_name like ? or email like ? limit ? offset ?"
	UserSearchCount = "select count(1) from user where name like ? or full_name like ? or email like ?"
	UserGetByID     = "select * from user where in ?"
)

/* sql about member */
const (
	MemberQuerybyID   = "select * from member where id = ? and deleted_at is null"
	MemberSingleQuery = "select * from member where resource_type = ? and  resource_id = ? and member_type= ?" +
		"and member_info = ? and deleted_at is null"
	MemberSingleDelete = "update member set deleted_at = CURRENT_TIMESTAMP where ID = ?"
	MemberSelectAll    = "select * from member where resource_type = ? and resource_id = ? and deleted_at is null"
)

/* sql about group */
const (
	GroupQueryByParentIDAndName = "select * from `group` where parent_id = ? and name = ? and deleted_at is null"
	GroupQueryByParentIDAndPath = "select * from `group` where parent_id = ? and path = ? and deleted_at is null"
	GroupDelete                 = "update `group` set deleted_at = CURRENT_TIMESTAMP where id = ?"
	GroupUpdateBasic            = "update `group` set name = ?, path = ?, description = ?, visibility_level = ? " +
		"where id = ?"
	GroupUpdateParentID           = "update `group` set parent_id = ? where id = ?"
	GroupQueryByID                = "select * from `group` where id = ? and deleted_at is null"
	GroupQueryByIDs               = "select * from `group` where id in ? and deleted_at is null"
	GroupQueryByPaths             = "select * from `group` where path in ? and deleted_at is null"
	GroupQueryByNameFuzzily       = "select * from `group` where name like ? and deleted_at is null"
	GroupUpdateTraversalIDs       = "update `group` set traversal_ids = ? where id = ? and deleted_at is null"
	GroupCountByParentID          = "select count(1) from `group` where parent_id = ? and deleted_at is null"
	GroupUpdateTraversalIDsPrefix = "update `group` set traversal_ids = replace(traversal_ids, ?, ?) " +
		"where traversal_ids like ? and deleted_at is null"
	GroupQueryByNameOrPathUnderParent = "select * from `group` where parent_id = ? " +
		"and (name = ? or path = ?) and deleted_at is null"
)

/* sql about application */
const (
	// ApplicationQueryByName ...
	ApplicationQueryByName            = "select * from application where name = ? and deleted_at is null"
	ApplicationQueryByNamesUnderGroup = "select * from application where group_id = ? and name in ? " +
		"and deleted_at is null"
	ApplicationDeleteByName = "update application set deleted_at = CURRENT_TIMESTAMP where name = ?"
)
