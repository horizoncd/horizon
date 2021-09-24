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
)

/* sql about group */
const (
	GroupQueryByParentIDAndName  = "select * from `group` where parent_id = ? and name = ? and deleted_at is null"
	GroupQueryByParentIDAndPath  = "select * from `group` where parent_id = ? and path = ? and deleted_at is null"
	GroupDelete                  = "update `group` set deleted_at = CURRENT_TIMESTAMP where id = ?"
	GroupUpdateBasic             = "update `group` set name = ?, path = ?, description = ?, visibility_level = ?"
	GroupQueryByID               = "select * from `group` where id = ? and deleted_at is null"
	GroupQueryByIDs              = "select * from `group` where id in ? and deleted_at is null"
	GroupQueryByIDsOrderByIDDesc = "select * from `group` where id in ? and deleted_at is null order by id desc"
	GroupQueryByPaths            = "select * from `group` where path in ? and deleted_at is null"
	GroupQueryByNameFuzzily      = "select * from `group` where name like ? and deleted_at is null"
	GroupUpdateTraversalIDs      = "update `group` set traversal_ids = ? where id = ? and deleted_at is null"
	GroupCountByParentID         = "select count(1) from `group` where parent_id = ? and deleted_at is null"
)
